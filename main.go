package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hewo233/house-system-backend/Init"
	"github.com/hewo233/house-system-backend/route"
	"log"
)

func main() {
	Init.AllInit()

	r := gin.Default()
	route.InitRoute(r)
	err := r.Run(":8080")
	if err != nil {
		log.Fatal("cannot start gin engine")
	}
}
