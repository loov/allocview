package main

import (
	"math/rand"
	"time"

	_ "loov.dev/allocview/attach"
)

func main() {
	total := 0
	var leak [][]byte
	for {
		alloc := 1000 + rand.Intn(1000)*10
		leak = append(leak, make([]byte, alloc))
		total += alloc

		if rand.Intn(10) > 8 {
			i := rand.Intn(len(leak))
			p := leak[i]
			leak = append(leak[:i], leak[i+1:]...)
			total -= len(p)
		}

		jitter := 30 + rand.Intn(30)
		time.Sleep(time.Duration(jitter) * time.Millisecond)
	}
}
