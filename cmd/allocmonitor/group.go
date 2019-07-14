package main

import "sync"

type Group struct {
	wait sync.WaitGroup
}

func (group *Group) Go(fn func()) {
	group.wait.Add(1)
	go func() {
		defer group.wait.Done()
		fn()
	}()
}

func (group *Group) Wait() {
	group.wait.Wait()
}
