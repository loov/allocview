package attach

import (
	"bufio"
	"encoding/binary"
	"os"
	"reflect"
	"runtime"
	"time"
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
	out := bufio.NewWriter(f)
	enc := EncodeWriter{out}

	out.Write([]byte("alloclog\x00"))
	enc.WriteString(exe)

	name, addr := Addr()
	enc.WriteString(name)
	enc.WriteUintptr(addr)

	enc.Flush()
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
		enc.WriteInt64(t.UnixNano())
		enc.WriteUint32(uint32(n))
	nextRecord:
		for _, rec := range records[:n] {
			enc.WriteInt64(rec.AllocBytes)
			enc.WriteInt64(rec.FreeBytes)
			enc.WriteInt64(rec.AllocObjects)
			enc.WriteInt64(rec.FreeObjects)
			for _, frame := range rec.Stack0 {
				enc.WriteUintptr(frame)
				if frame == 0 {
					continue nextRecord
				}
			}
			enc.WriteUintptr(0)
		}
		enc.Flush()
		f.Sync()
	}
}

type EncodeWriter struct {
	*bufio.Writer
}

func (w *EncodeWriter) WriteUint64(v uint64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], v)
	w.Write(data[:])
}

func (w *EncodeWriter) WriteUintptr(v uintptr) {
	w.WriteUint64(uint64(v))
}

func (w *EncodeWriter) WriteInt64(v int64) {
	w.WriteUint64(uint64(v))
}

func (w *EncodeWriter) WriteUint32(v uint32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], v)
	w.Write(data[:])
}

func (w *EncodeWriter) WriteString(s string) {
	w.WriteUint32(uint32(len(s)))
	w.Write([]byte(s))
}
