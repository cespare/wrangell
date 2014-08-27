package main

import (
	"fmt"
	"git"
)

func main() {
	repo, err := git.Open(".")
	if err != nil {
		panic(err)
	}
	for _, p := range []string{
		"master",
		"origin/master",
		"14242a95a1d2403128a1a9fbe9f5efec43d04fac",
		"a9d7100fe1a37926a0a9cb05992e6eb3cdcb4f0d",
		//"a9d7",
	} {
		commit, err := repo.Ref(p)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s: %s\n", p, commit.SHA)
		}
	}
}
