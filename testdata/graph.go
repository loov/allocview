package main

import (
	"runtime"
	"time"
)

type Node struct {
	Data  []byte
	Links []*Node
}

func N(size int) *Node {
	return &Node{Data: make([]byte, size)}
}

var a, b, c, d, e, x *Node

func main() {
	a = N(1 << 20)
	b = N(2 << 20)
	c = N(3 << 20)
	d = N(4 << 20)
	e = N(5 << 20)

	a.Links = []*Node{b, c}
	b.Links = []*Node{d, e, c}
	c.Links = []*Node{e}
	e.Links = []*Node{a, b}

	for i := 0; i < 10; i++ {
		x = N(6 << 20)
		time.Sleep(1 * time.Second)
	}

	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	runtime.KeepAlive(c)
	runtime.KeepAlive(d)
	runtime.KeepAlive(e)
	runtime.KeepAlive(x)

	runtime.GC()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
