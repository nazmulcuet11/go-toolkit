package main

import "github.com/nazmulcuet11/go-toolkit/toolkit"

func main() {
	var t *toolkit.Tools
	t.CreateDirIfNotExists("./test-dir/subdir")
}
