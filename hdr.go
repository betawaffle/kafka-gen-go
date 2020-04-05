package main

import (
	"sync"

	"github.com/betawaffle/kafka-gen-go/codegen"
	"github.com/betawaffle/kafka-gen-go/schema"
)

type hdrGenerator struct {
	mu  sync.Mutex
	req *schema.MessageData
	res *schema.MessageData
}

func (g *hdrGenerator) addMessage(m *schema.MessageData) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	var ok bool
	switch m.Name {
	case "RequestHeader":
		if g.req != nil {
			panic("request header conflict")
		}
		g.req, ok = m, g.res != nil
	case "ResponseHeader":
		if g.res != nil {
			panic("response header conflict")
		}
		g.res, ok = m, g.req != nil
	default:
		panic("BUG: unexpected header name: " + m.Name)
	}
	return ok
}

func (g *hdrGenerator) getFileName() string {
	return "headers_gen.go"
}

func (g *hdrGenerator) run(w *codegen.File) {
	genMessage(w, g.req)
	genMessage(w, g.res)
}
