package main

import (
	"log"

	"kanalegeleri/go-app/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
