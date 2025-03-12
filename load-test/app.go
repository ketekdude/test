package main

import (
	"log"
	"test/load-test/handler/rest"
)

func main() {
	rest.Init()
	log.Println("Server started on :8080")
}
