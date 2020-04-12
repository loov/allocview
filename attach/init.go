package attach

import (
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/loov/allocview/internal/packet"
)

// Addr returns the address of this func.
// go:noinline
func Addr() (string, uintptr) {
	addr := reflect.ValueOf(Addr).Pointer()
	fn := runtime.FuncForPC(addr)
	return fn.Name(), addr
}

func init() {
	sockPath := os.Getenv("ALLOCLOG")
	if sockPath == "" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	f, err := os.Create(sockPath)
	if err != nil {
		panic(err)
	}

	runtime.MemProfileRate = 1
	monitor(exe, f)
}

func monitor(exe string, f *os.File) {
	enc := packet.NewEncoder(1 << 20)

	enc.String("alloclog\x00")
	enc.String(exe)

	name, addr := Addr()
	enc.String(name)
	enc.Uintptr(addr)

	f.Write(enc.LengthAndBytes())
	f.Sync()

	tick := time.NewTicker(time.Second / 10)
	records := make([]runtime.MemProfileRecord, 1000)
	for t := range tick.C {
	tryagain:
		n, ok := runtime.MemProfile(records, false)
		if !ok {
			records = make([]runtime.MemProfileRecord, n+n/3)
			goto tryagain
		}

		enc.Reset()
		enc.Int64(t.UnixNano())
		enc.Uint32(uint32(n))
	nextRecord:
		for _, rec := range records[:n] {
			enc.Int64(rec.AllocBytes)
			enc.Int64(rec.FreeBytes)
			enc.Int64(rec.AllocObjects)
			enc.Int64(rec.FreeObjects)
			for _, frame := range rec.Stack0 {
				enc.Uintptr(frame)
				if frame == 0 {
					continue nextRecord
				}
			}
			enc.Uintptr(0)
		}

		f.Write(enc.LengthAndBytes())
		f.Sync()
	}
}
