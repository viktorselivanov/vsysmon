package main

import (
	"flag"
	"log"
)

var port = flag.Int("p", 50051, "port")

func main() {
	flag.Parse()

	if err := RunClient(); err != nil {
		log.Fatal(err) // завершает программу при ошибке
	}
}
