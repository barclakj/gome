package main

import (
	"log"
	"sync"

	"realizr.io/gome/ctrl"
)

var wg sync.WaitGroup

func main() {
	ctrl := ctrl.LogEntryController{}
	ctrl.Init(&wg)

	f := dostuff()
	defer f(10)
	log.Printf("Hello world!\n")

	wg.Wait()
}

func dostuff() func(v int) {
	return func(v int) {
		log.Printf("Called with %d\n", v)
	}
}
