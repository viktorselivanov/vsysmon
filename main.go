package main

import (
	"flag"
	"vsysmon/collectors"
	"vsysmon/report"
	"vsysmon/ring"
)

var (
	N       = flag.Int("n", 5, "report interval in seconds (1-60)")     // переменная для выдачи информации каждые N секунд
	M       = flag.Int("m", 15, "aggregation window in seconds (1-60)") // переменная для усреднения информации за последние M секунд.
	verbose = flag.Bool("v", false, "verbose")                          // переменная для включения логирования в терминал
	port    = flag.Int("p", 50051, "TCP port to listen on")             // переменная для выбора порта
)

func main() {

	flag.Parse()

	InitVerbose(*verbose) // включение/выключение логгирования

	validateFlags() // проверка флагов

	cfg := loadConf() // загрузка конфига

	ring.RingInit(*M) // инициализация кольцевого буффера

	go report.StartGRPC(*port) // старт GRPC сервера

	done := make(chan struct{}) // создаём небуферизированный канал

	pipeline := collectors.BuildPipeline(cfg) // сборка пайплайна (используется для включения/выключения функций)

	collectors.StartCollector(done, pipeline) // запуск коллектора

	go report.Reporter(cfg, *verbose, *N) // выдача информации

	select {} // бесконечный цикл
}
