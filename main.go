package main

import (
	"sync"

	"realizr.io/gome/ctrl"
	"realizr.io/gome/rest"
)

var wg sync.WaitGroup

func main() {
	ctrl := ctrl.LogEntryController{}
	ctrl.Init(&wg)

	go rest.StartWebServer(&ctrl, &wg)

	wg.Wait()
}
