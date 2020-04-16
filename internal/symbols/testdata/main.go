package main

import (
	"fmt"
	"runtime"

	"loov.dev/allocview/attach"
)

//go:noinline
func main() {
	hello()
}

//go:noinline
func hello() {
	world()
}

//go:noinline
func world() {
	var pcs [10]uintptr
	n := runtime.Callers(1, pcs[:])
	for _, pc := range pcs[:n] {
		fmt.Println(pc)
	}
}

func init() {
	name, addr := attach.Addr()
	fmt.Println(name)
	fmt.Println(addr)
}
