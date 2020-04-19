package attach

import (
	"net"
	"os"
	"reflect"
	"runtime"
	"time"

	"loov.dev/allocview/internal/packet"
)

// Addr returns the address of this func.
// go:noinline
func Addr() (string, uintptr) {
	addr := reflect.ValueOf(Addr).Pointer()
	fn := runtime.FuncForPC(addr)
	return fn.Name(), addr
}

func init() {
	sockPath := os.Getenv("ALLOCLOGSOCK")
	if sockPath == "" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	sockAddr, err := net.ResolveUnixAddr("unix", sockPath)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUnix("unix", nil, sockAddr)
	if err != nil {
		panic(err)
	}

	runtime.MemProfileRate = 1

	err = monitor(exe, conn)
	if err != nil {
		panic(err)
	}
}

func monitor(exe string, conn *net.UnixConn) error {
	enc := packet.NewEncoder(1 << 20)

	enc.String("alloclog")
	enc.String(exe)

	name, addr := Addr()
	enc.String(name)
	enc.Uintptr(addr)

	if _, err := conn.Write(enc.LengthAndBytes()); err != nil {
		_ = conn.Close()
		return err
	}

	go func() {
		defer conn.Close()

		tick := time.NewTicker(time.Second / 10)
		records := make([]runtime.MemProfileRecord, 1000)
		for t := range tick.C {
			// TODO: figure out a better way to do this
			// runtime.GC forces mem profile to be updated
			runtime.GC()
		tryagain:
			n, ok := runtime.MemProfile(records, true)
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

			if _, err := conn.Write(enc.LengthAndBytes()); err != nil {
				panic(err)
			}
		}
	}()

	return nil
}
