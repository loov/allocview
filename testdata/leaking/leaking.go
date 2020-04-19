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
		alloc := 10000 + rand.Intn(1000)*10
		mem := make([]byte, alloc)
		leak = append(leak, mem)
		total += len(mem)

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
