package main

import (
	"log"

	"github.com/nazmulcuet11/go-toolkit/toolkit"
)

func main() {
	var tools toolkit.Tools
	slug, err := tools.Slugify("NOW!!? is the time 123")
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(slug)
}
