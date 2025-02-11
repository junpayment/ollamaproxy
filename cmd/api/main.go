package main

import (
	"log"

	"github.com/junpayment/ollamaproxy/cmd/api/di"
)

func main() {
	api := di.InitAPI()
	err := api.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
