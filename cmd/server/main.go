package main

import (
	"log"

	"github.com/code-practice-archives/api-demo/internal/server"
)

func main() {
	r := server.New()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
