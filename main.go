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
	for _, p := range []string{"master", "origin/master", "a9d7100fe1a37926a0a9cb05992e6eb3cdcb4f0d"} {
		commit, err := repo.Ref(p)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: %s\n", p, commit.SHA)
	}
}
