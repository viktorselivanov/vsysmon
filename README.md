# vsysmon
**vsysmon** — лёгкий агент мониторинга системы, написанный на Go без использования сторонних пакетов.  
Он собирает системные метрики, хранит последние сэмплы в кольцевом буфере и отображает агрегированный снимок в консоли.

---

## Возможности
Работа с OC Linux и Windows

### Сбор метрик

Для Linux 

- CPU (user / system / idle)
- Load average
- Disk IO (TPS, KB/s)
- Filesystems (statfs)
- TCP состояния (ESTABLISHED, LISTEN и др.)
- Listening sockets (порт, PID, пользователь, команда)
- Top Talkers:
  - по протоколам (TCP / UDP / ICMP)
  - по потокам (src → dst)

Для Windows 

- CPU (user / system / idle)

Модульная структура для удобной интеграции новых возможностей.
Единая реализация отображения информации для клиента и сервера через рендеры.
Удобная запуска и отображения метрик через конфиг файл config.json*

*   В случае отсутсвия config.json при запуске агента используется default config со всеми метриками
    Для клиента всегда используется default config, в случае отключения метрик будут отображены нулевые значения соответственно


{
  "linux":{
  "collect_load": true,
  "collect_cpu": true,
  "collect_disk": true,
  "collect_fs": true,
  "collect_tcp_states": true,
  "collect_listen": true,
  "collect_top_talkers": true
  },
  "windows":{
  "collect_load": true,
  "collect_cpu": true
  }
}

### Архитектура
- Пайплайн коллекторов
- Потокобезопасный кольцевой буфер
- Терминальные рендереры (одна секция = один рендерер)
- Юнит-тесты с моками `/proc`
- Интеграционные тесты под Linux (`-tags=integration`)

---

## Структура проекта

vsysmon/
├── bin/ # собранные бинарники
├── client/ # клиент gRPC
├── collectors/ # коллекторы и пайплайн
├── config/ # конфигурация
├── model/ # структуры Sample / Snapshot
├── proto/ # protobuf 
├── report/ # агрегация и формирование snapshot,  grpc сервер
├── ring/ # кольцевой буфер и последний snapshot
├── terminal/ # терминальные рендереры
├── ring/ # кольцевой буфер
├── model/ # структуры Sample / Snapshot
├── init.go # инициализация
├── main.go # агент
└── README.md

---

## Запуск

Агент

$cd bin 
$./vsysmon -v

Опционально можно настраивать m, n и порт. А так же включать отображение метрик без использования клиента
Usage of ./vsysmon:
  -m int
        aggregation window in seconds (1-60) (default 15)
  -n int
        report interval in seconds (1-60) (default 5)
  -p int
        TCP port to listen on (default 50051)
  -v    verbose


Клиент 
$cd bin 
$./client

Usage of ./bin/client:
  -p int
        port (default 50051)

Для Windows c испоьзованием файлов с расширением .exe

---

## Пример отображения

======== SYSTEM SNAPSHOT ========

LOAD AVERAGE
-----------------------------
Load avg: 2.66

CPU USAGE (%)
-----------------------------
User           2.08
System         4.81
Idle          91.84


DISK IO
-----------------------------
TPS              7.87
KB/s            80.53


FILESYSTEMS
--------------------------------------------------------------------------------
FS                             MOUNT             USED(MB)        %      INODE        %
/dev/mapper/vgubuntu-root      /                   773504     82.7    2928383      4.8
/dev/nvme12345                 /boot                  303     18.5        322      0.3


TCP STATES
-----------------------------
STATE              COUNT
  TIME_WAIT: 120
  LISTEN: 288
  ESTABLISHED: 239


TOP TALKERS — BY PROTOCOL
--------------------------------------------
PROTO         BYTES/s        %
TCP          13550521   100.0%
UDP                 0     0.0%


TOP TALKERS — BY FLOW (BPS)
---------------------------------------------------------------------
SRC                    -> DST                    PROTO           BPS
127.0.0.1:45273        -> 127.0.0.2:45989        TCP       1360961
127.0.0.1:45273        -> 127.0.0.4:47461        TCP       1360960
127.0.0.1:45273        -> 127.0.0.5:32941        TCP       1359106
127.0.0.5:32941        -> 127.0.0.1:45273        TCP       1359105
127.0.0.1:45273        -> 0.0.0.0:0              TCP       1359103
127.0.0.1:45273        -> 127.0.0.3:46541        TCP       1357811
127.0.0.3:46541        -> 127.0.0.1:45273        TCP       1357810
127.0.0.2:45989        -> 127.0.0.1:45273        TCP       1350829
127.0.0.4:47461        -> 127.0.0.1:45273        TCP       1348897
172.16.16.17:49228     -> 172.16.16.254:443      TCP       1335939


LISTENING SOCKETS
-----------------------------------------------------------
P      PORT   PID      USER       CMD
UDP      5353   pc         16431  opera
UDP      5353   pc         16431  opera
UDP      5353   pc         16431  opera
UDP      5353   pc         16431  opera
TCP6     50051  pc         260821 vsysmon