package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/betawaffle/kafka-gen-go/codegen"
	"github.com/betawaffle/kafka-gen-go/schema"
	"github.com/iancoleman/strcase"
)

type apiGenerator struct {
	mu  sync.Mutex
	req *schema.MessageData
	res *schema.MessageData
}

func (g *apiGenerator) addMessage(m *schema.MessageData) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	var ok bool
	switch m.Type {
	case "request":
		if g.req != nil {
			panic("request conflict")
		}
		g.req, ok = m, g.res != nil
	case "response":
		if g.res != nil {
			panic("response conflict")
		}
		g.res, ok = m, g.req != nil
	default:
		panic("BUG: unexpected message type: " + m.Type)
	}
	return ok
}

func (g *apiGenerator) getFileName() string {
	req := strings.TrimSuffix(g.req.Name, "Request")
	res := strings.TrimSuffix(g.res.Name, "Response")
	if req != res {
		panic(fmt.Errorf("mismatched req/res names; %s != %s", req, res))
	}
	return strcase.ToSnake(req) + "_gen.go"
}

func (g *apiGenerator) run(w *codegen.File) {
	genMessage(w, g.req)
	genMessage(w, g.res)
}
