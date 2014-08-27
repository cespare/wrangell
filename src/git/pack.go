package git

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/edsrzf/mmap-go"
)

// Functions for reading git's pack format.
// References:
// https://www.kernel.org/pub/software/scm/git/docs/technical/pack-format.txt
// http://schacon.github.io/gitbook/7_the_packfile.html

// TODO: Stop using os.IsNotExist in so many places, and introduce my own concept of "does not exist" errors
// so all the object finding code can distinguish between "no such object" and "something bad happened".

func (r *Repo) packedObjectBySHA(sha *SHA) (*Object, error) {
	// TODO: Handle unindexed packfiles (maybe by getting get to fix them first). (This is an unexpected case
	// though.)
	idxFiles, err := filepath.Glob(filepath.Join(r.gitDir, "objects", "pack", "pack-*.idx"))
	if err != nil {
		return nil, err
	}
	for _, idxFile := range idxFiles {
		off, err := r.packIdxLookup(sha, idxFile)
		if err == nil {
			packfileName := strings.TrimSuffix(idxFile, ".idx") + ".pack"
			return r.packFileLookup(sha, packfileName, off)
		}
		if err != objNotInIdxErr {
			return nil, err
		}
	}
	return nil, os.ErrNotExist
}

var (
	objNotInIdxErr    = errors.New("object was not located in the index")
	packIdxHeaderErr  = errors.New("pack index file did not have a recognized header for version 2+")
	packIdxVersionErr = errors.New("pack index file was not version 2.")
)

// packIdxLookup finds a SHA in a pack index. It returns the pack offset if it locates the entry.
func (r *Repo) packIdxLookup(sha *SHA, filename string) (uint64, error) {
	f, m, err := openMap(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	defer m.Unmap()

	// Check the header
	if !bytes.Equal(m[:4], []byte("\377tOc")) {
		return 0, packIdxHeaderErr
	}
	// Check the index version number
	if readUint32(m[4:], 0) != 2 {
		return 0, packIdxVersionErr
	}
	// Get the bounds for the binary search from the first-level fanout table.
	fanout := m[8:]
	shaSlice := sha.Bytes()
	b0 := shaSlice[0]
	var lo int
	if b0 > 0 {
		lo = int(readUint32(fanout, int(b0-1)))
	}
	hi := int(readUint32(fanout, int(b0)))
	numObjects := int(readUint32(fanout, 255))
	// Binary search to find our sha.
	// We could call sort.Search here but it's more trouble than it's worth.
	shaBase := 4 + 4 + 256*4 // magic + header + fan-out table
	shaIdx := -1
binSearch:
	for lo <= hi {
		mid := (lo + hi) / 2
		start := shaBase + int(mid)*20
		test := m[start : start+20]
		switch bytes.Compare(shaSlice, test) {
		case 0:
			shaIdx = mid
			break binSearch
		case -1:
			hi = mid - 1
		case 1:
			lo = mid + 1
		}
	}
	if shaIdx == -1 {
		return 0, objNotInIdxErr
	}
	// Now get the pack file offset from the list of offsets following the list of object names.
	offBase := shaBase + numObjects*20 + numObjects*4 // skip over the SHAs and the CRC32 sums
	off := readUint32(m[offBase:], shaIdx)
	msb, off := splitMSB32(off)
	if !msb {
		// The 31 LSBs of off are the pack file offset.
		return uint64(off), nil
	}
	// Look up the offset in the large-offset table.
	largeOffBase := offBase + numObjects*4
	largeOffIdx := int(off)
	return readUint64(m[largeOffBase:], largeOffIdx), nil
}

var (
	packHeaderErr     = errors.New("packfile did not have a recognized header for version 2+")
	packVersionErr    = errors.New("packfile was not version 2.")
	packBadObjTypeErr = errors.New("encountered unknown object type")
)

func (r *Repo) packFileLookup(sha *SHA, filename string, off uint64) (*Object, error) {
	f, m, err := openMap(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	defer m.Unmap()

	// Check the header
	if string(m[:4]) != "PACK" {
		return nil, packHeaderErr
	}
	if readUint32(m, 1) != 2 {
		return nil, packVersionErr
	}

	for {
		// Read the type and size
		i := int(off)
		msb, part := splitMSB8(m[i])
		size := int64(part & 0xF)    // bits 0 - 3
		typ := ObjectType(part >> 4) // bits 4, 5, 6
		i++
		for sizeBits := uint(4); msb; i, sizeBits = i+1, sizeBits+7 {
			msb, part = splitMSB8(m[i])
			size += int64(part) << sizeBits
		}
		switch typ {
		case TypeCommit, TypeTree, TypeBlob, TypeTag:
			return &Object{
				SHA:  sha,
				Type: typ,
			}, nil
		case TypeOfsDelta:
			msb, part := splitMSB8(m[i])
			delta := uint64(part)
			var extra uint64
			for n := uint(1); msb; n++ {
				msb, part = splitMSB8(m[i+int(n)])
				delta = delta<<7 + uint64(part)
				extra += 1 << (7 * n)
			}
			delta += extra
			if delta > off {
				return nil, errors.New("corrupt packfile: negative offset delta too large")
			}
			off -= delta
		case TypeRefDelta:
			panic("unimplemented")
		default:
			return nil, packBadObjTypeErr
		}
	}
}

func openMap(filename string) (*os.File, mmap.MMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	m, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, m, err
}

// splitMSB8 returns whether the MSB is set and the remaining 7-bit integer as a byte.
func splitMSB8(b byte) (msb bool, rest byte) {
	const mask byte = 1 << 7
	msb = b&mask > 0
	rest = b & ^mask
	return
}

// splitMSB32 returns whether the MSB is set and the remaining 31-bit integer as a uint32.
func splitMSB32(n uint32) (msb bool, rest uint32) {
	const mask uint32 = 1 << 31
	msb = n&mask > 0
	rest = n & ^mask
	return
}

// readUint32 considers b as a sequence of 4-byte chunks, and reads one big-endian uint32 from the ith chunk.
// (So readUint32(b, 4) considers the bytes in b[16:20]).
func readUint32(b []byte, i int) uint32 {
	base := i * 4
	return uint32(b[base+3]) | uint32(b[base+2])<<8 | uint32(b[base+1])<<16 | uint32(b[base])<<24
}

// readUint64 considers b as a sequence of 8-byte chunks, and reads one big-endian uint64 from the ith chunk.
func readUint64(b []byte, i int) uint64 {
	base := i * 8
	return uint64(b[base+7]) | uint64(b[base+6])<<8 | uint64(b[base+5])<<16 | uint64(b[base+4])<<24 |
		uint64(b[base+3])<<32 | uint64(b[base+2])<<40 | uint64(b[base+1])<<48 | uint64(b[base+0])<<56
}
