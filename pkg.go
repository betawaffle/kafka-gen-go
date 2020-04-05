package main

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/betawaffle/kafka-gen-go/codegen"
	"github.com/betawaffle/kafka-gen-go/schema"
)

type genImpl interface {
	addMessage(*schema.MessageData) bool
	getFileName() string
	run(*codegen.File)
}

type pkgGenerator struct {
	wg sync.WaitGroup
	mu sync.Mutex

	dst string
	hdr hdrGenerator
	api map[int16]*apiGenerator
}

func (g *pkgGenerator) addFile(src string) {
	defer g.wg.Done()

	m, err := schema.ReadMessageData(src)
	if err != nil {
		log.Print(err)
		return
	}

	var impl genImpl
	switch m.Type {
	case "header":
		impl = &g.hdr
	case "request", "response":
		impl = g.getAPI(*m.ApiKey)
	default:
		panic("unexpected message type: " + m.Type)
	}
	if !impl.addMessage(m) {
		return
	}
	f := codegen.NewFile(filepath.Join(g.dst, impl.getFileName()))
	impl.run(f)

	if err = f.Flush(); err != nil {
		log.Print(err)
	}
}

func (g *pkgGenerator) getAPI(key int16) *apiGenerator {
	g.mu.Lock()
	defer g.mu.Unlock()

	if a, ok := g.api[key]; ok {
		return a
	}
	a := &apiGenerator{}
	g.api[key] = a
	return a
}

func (g *pkgGenerator) finish() {
}
