package main

import (
	"log"

	"github.com/l402-protocol/go-example/gateway"
)

func main() {
	g := gateway.NewGateway()
	log.Println("Starting gateway server on :8081")
	if err := g.Start(":8081"); err != nil {
		log.Fatal(err)
	}
}
