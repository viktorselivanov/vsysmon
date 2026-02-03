package main

import (
	"flag"
	"fmt"
	"os"
	"vsysmon/config"
)

var (
	VPrintln = func(...any) {}
	VPrintf  = func(string, ...any) {}
)

func InitVerbose(on bool) { // включение/выключение логгирования
	if on {
		VPrintln = func(a ...any) { fmt.Println(a...) }
		VPrintf = func(f string, a ...any) { fmt.Printf(f, a...) }
	}
}

func validateFlags() { // проверка флагов

	if *N < 1 || *N > 60 || *M < 1 || *M > 60 {
		die("n and m must be in range 1..60")
	}
	if *M < *N {
		die("m must be >= n")
	}
	if *port < 1 || *port > 65535 {
		die("invalid port: (must be 1..65535)")
	}
	VPrintln("Values -> N:", *N, "M:", *M, "| Port:", *port) // вывод начальных данных
}

func die(msg string) { // завершаем в случае ошибки
	fmt.Println("Error:", msg, "")
	flag.Usage()
	os.Exit(1)
}

func loadConf() config.Config {

	cfg, err := config.LoadConfig("config.json") // загрузка конфига, если таковой имеется и вывод справочной информации
	if err != nil {
		cfg = config.DefaultConfig()
		VPrintf("Configuration not found. If you want to change the properties, specify them in the config.json file.\n")
		VPrintf("Default properties will be used:\n%+v\n", cfg)
	} else {
		VPrintf("Loaded config: %+v\n", cfg)
	}
	return cfg
}
