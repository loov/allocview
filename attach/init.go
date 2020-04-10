package attach

import (
	"reflect"
	"runtime"
)

// Addr returns the address of this func.
// go:noinline
func Addr() (string, uintptr) {
	addr := reflect.ValueOf(Addr).Pointer()
	fn := runtime.FuncForPC(addr)
	return fn.Name(), addr
}
