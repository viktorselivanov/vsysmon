//go:build windows
// +build windows

package collectors

import (
	"fmt"
	"time"
	"unsafe"

	model "vsysmon/internal/model"

	"golang.org/x/sys/windows"
)

var (
	modpdh                           = windows.NewLazySystemDLL("pdh.dll")           // доступ к системным счетчикам
	procPdhOpenQueryW                = modpdh.NewProc("PdhOpenQueryW")               // Создаёт новый запрос для сбора счетчиков
	procPdhAddEnglishCounterW        = modpdh.NewProc("PdhAddEnglishCounterW")       // Добавляет конкретный счетчик в запрос
	procPdhCollectQueryData          = modpdh.NewProc("PdhCollectQueryData")         // Собирает текущие значения всех счетчиков в запросе
	procPdhGetFormattedValue         = modpdh.NewProc("PdhGetFormattedCounterValue") // Получает отформатированное значение конкретного счетчика
	procPdhCloseQuery                = modpdh.NewProc("PdhCloseQuery")               // Закрывает запрос и освобождает ресурсы
	PDH_FMT_DOUBLE            uint32 = 0x00000200                                    // Определяет, что значение счётчика нужно вернуть как double. PDH ждёт DWORD (32 бита)
)

type PDH_FMT_COUNTERVALUE_DOUBLE struct { // структура для Performance Data Helper
	CStatus     uint32
	DoubleValue float64
}

func must(ret uintptr, name string) { // Любое ненулевое значение — это код ошибки
	if ret != 0 {
		panic(fmt.Sprintf("%s failed: 0x%X", name, ret))
	}
}

// cpuStat для хранения предыдущих значений
type cpuStat struct {
	user, sys, idle float64
}

// CPUCollector Windows-подобный интерфейс с prev
type CPUCollector struct {
	prev cpuStat
	init bool

	cUser, cSys, cIdle windows.Handle
	query              windows.Handle // handle запроса
}

func (c *CPUCollector) Name() string { return "cpu" }

// инициализация PDH счётчиков
func (c *CPUCollector) initCounters() {
	ret, _, _ := procPdhOpenQueryW.Call(0, 0, uintptr(unsafe.Pointer(&c.query))) // локальная система, user data не используем
	must(ret, "PdhOpenQueryW")

	add := func(path string, handle *windows.Handle) { // унифицируем что бы не дублировать код
		p, _ := windows.UTF16PtrFromString(path)     // конвертация строки
		ret, _, _ := procPdhAddEnglishCounterW.Call( // добавление счётчика
			uintptr(c.query),
			uintptr(unsafe.Pointer(p)),
			0,
			uintptr(unsafe.Pointer(handle)),
		)
		must(ret, "PdhAddEnglishCounterW "+path)
	}

	add(`\Processor(_Total)\% User Time`, &c.cUser)
	add(`\Processor(_Total)\% Privileged Time`, &c.cSys)
	add(`\Processor(_Total)\% Idle Time`, &c.cIdle)

	ret, _, _ = procPdhCollectQueryData.Call(uintptr(c.query)) // первичный вызов
	must(ret, "PdhCollectQueryData(init)")
	c.init = true
}

// Collect собирает значения CPU и считает дельту
func (c *CPUCollector) Collect(s *model.Sample) {
	if !c.init {
		c.initCounters()
		time.Sleep(1 * time.Second) // короткая пауза для корректного измерения
	}

	ret, _, _ := procPdhCollectQueryData.Call(uintptr(c.query))
	must(ret, "PdhCollectQueryData(read)") // собирает текущее состояние всех счётчиков

	read := func(h windows.Handle) float64 {
		var v PDH_FMT_COUNTERVALUE_DOUBLE
		ret, _, _ := procPdhGetFormattedValue.Call(
			uintptr(h),
			uintptr(PDH_FMT_DOUBLE),
			0,
			uintptr(unsafe.Pointer(&v)), // для передачи указателя в syscall
		)
		must(ret, "PdhGetFormattedCounterValue")
		return v.DoubleValue
	}

	cur := cpuStat{
		user: read(c.cUser),
		sys:  read(c.cSys),
		idle: read(c.cIdle),
	}

	if !c.init {
		c.prev = cur
		c.init = true
		return
	}

	// вычисляем дельту как в Linux (хотя значения уже в процентах)
	s.CPUUser = cur.user
	s.CPUSys = cur.sys
	s.CPUIdle = cur.idle
	c.prev = cur
}

// Close закрывает PDH query
func (c *CPUCollector) Close() {
	if c.query != 0 {
		procPdhCloseQuery.Call(uintptr(c.query))
		c.query = 0
	}
}
