package cron

import (
	"sort"
	"testing"
)

// One []int each for minutes, hours, ...
// must be in sorted order
// nil == '*'
type testSchedule [5][]int

type parseTestCase struct {
	expr   string
	parsed testSchedule
}

func TestParse(t *testing.T) {
	for _, testCase := range []parseTestCase{
		{"* * * * *", testSchedule{nil, nil, nil, nil, nil}},
		{"0 0 1 1 0", testSchedule{{0}, {0}, {0}, {0}, {0}}},
		{"2,3 * * * *", testSchedule{{2, 3}, nil, nil, nil, nil}},
		{"2-5 * * * *", testSchedule{{2, 3, 4, 5}, nil, nil, nil, nil}},
		{"1,3-5 * * * *", testSchedule{{1, 3, 4, 5}, nil, nil, nil, nil}},
		{"* 21-3 * * *", testSchedule{nil, {0, 1, 2, 3, 21, 22, 23}, nil, nil, nil}},
		{"* * * JAN *", testSchedule{nil, nil, nil, {0}, nil}},
		{"* * * Janua *", testSchedule{nil, nil, nil, {0}, nil}},
		{"* * * APR-JUL *", testSchedule{nil, nil, nil, {3, 4, 5, 6}, nil}},
		{"* * * * MON,WED", testSchedule{nil, nil, nil, nil, {1, 3}}},
		{"* */6 * * *", testSchedule{nil, {0, 6, 12, 18}, nil, nil, nil}},
		{"* 6-10/2 * * *", testSchedule{nil, {6, 8, 10}, nil, nil, nil}},
		{"@monthly", testSchedule{{0}, {0}, {0}, nil, nil}},
		{"@weekly", testSchedule{{0}, {0}, nil, nil, {0}}},
		{"@daily", testSchedule{{0}, {0}, nil, nil, nil}},
		{"@hourly", testSchedule{{0}, nil, nil, nil, nil}},
	} {
		s, err := Parse(testCase.expr)
		if err != nil {
			t.Fatal(err)
		}
		for i, size := range fieldSizes {
			set := testCase.parsed[i]
			if set == nil {
				for j := 0; j < size; j++ {
					set = append(set, j)
				}
			}
			for j := 0; j < size; j++ {
				setInS := s.isSet(fieldOffsets[i] + j)
				setInTestCase := true
				index := sort.SearchInts(set, j)
				if index == len(set) || set[index] != j {
					setInTestCase = false
				}
				if setInS != setInTestCase {
					t.Fatalf("got %v; want %v\n", s, testCase.parsed)
				}
			}
		}
	}
}

func TestParseFail(t *testing.T) {
	for _, expr := range []string{
		"* * * *",
		"-1 * * * *",
		"60 * * * *",
		"* 24 * * *",
		"* * 0 * *",
		"* * 32 * *",
		"* * * 0 *",
		"* * * 13 *",
		"* * * J *",
		"* * * foo *",
		"* * * * 7",
		"1 - 3 * * * *",
		"1-3-7 * * * *",
		"1/3/7 * * * *",
		"@foobar",
	} {
		if _, err := Parse(expr); err == nil {
			t.Fatalf("Parse accepted %q, but it is invalid", expr)
		}
	}
}
