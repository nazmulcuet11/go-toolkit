package main

import (
	"fmt"

	"github.com/nazmulcuet11/go-toolkit/toolkit"
)

func main() {
	tools := toolkit.Tools{}
	fmt.Println(tools.RandomString(32))
}
