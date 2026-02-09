package main

import (
	"flag"
	"log"
)

var (
	port = flag.Int("p", 50051, "port")
	N    = flag.Int("n", 5, "report interval in seconds (1-60)")     // переменная для выдачи информации каждые N секунд
	M    = flag.Int("m", 15, "aggregation window in seconds (1-60)") // переменная для усреднения за последние M секунд.
)

func main() {
	flag.Parse()

	if err := RunClient(); err != nil {
		log.Fatal(err) // завершает программу при ошибке
	}
}
