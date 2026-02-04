package main

import (
	"flag"
	"vsysmon/internal/collectors"
	"vsysmon/internal/report"
	"vsysmon/internal/ring"
)

var (
	N       = flag.Int("n", 5, "report interval in seconds (1-60)")     // переменная для выдачи информации каждые N секунд
	M       = flag.Int("m", 15, "aggregation window in seconds (1-60)") // переменная для усреднения за последние M секунд.
	verbose = flag.Bool("v", false, "verbose")                          // переменная для включения логирования в терминал
	port    = flag.Int("p", 50051, "TCP port to listen on")             // переменная для выбора порта
	cpath   = flag.String("c", "./config.json", "path to config")       // переменная для выбора пути к конфигу
)

func main() {
	flag.Parse()

	InitVerbose(*verbose) // включение/выключение логгирования

	validateFlags() // проверка флагов

	cfg := loadConf(*cpath) // загрузка конфига

	ring.Init(*M) // инициализация кольцевого буффера

	go report.StartGRPC(*port) // старт GRPC сервера

	done := make(chan struct{}) // создаём небуферизированный канал

	pipeline := collectors.BuildPipeline(cfg) // сборка пайплайна (используется для включения/выключения функций)

	collectors.StartCollector(done, pipeline) // запуск коллектора

	go report.Reporter(cfg, *verbose, *N) // выдача информации

	select {} // бесконечный цикл
}
