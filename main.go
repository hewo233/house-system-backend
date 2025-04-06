package main

import (
	"github.com/hewo233/house-system-backend/Init"
	"github.com/hewo233/house-system-backend/route"
	"log"
)

func main() {
	Init.AllInit()

	route.InitRoute()

	err := route.R.Run(":8080")
	if err != nil {
		log.Fatal("cannot start gin engine")
	}
}
