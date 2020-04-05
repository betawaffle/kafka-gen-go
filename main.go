package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	g := &pkgGenerator{
		dst: os.Args[1],
		api: make(map[int16]*apiGenerator),
	}
	for _, src := range os.Args[2:] {
		g.wg.Add(1)
		go g.addFile(src)
	}
	g.wg.Wait()
	g.finish()
}
