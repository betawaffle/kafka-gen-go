package main

import (
	"strconv"

	"github.com/betawaffle/kafka-gen-go/codegen"
	"github.com/betawaffle/kafka-gen-go/schema"
	"github.com/iancoleman/strcase"
)

func begMethod(w *codegen.File, recv, name, args, rets string) {
	w.WriteString("func (m *")
	w.WriteString(recv)
	w.WriteString(") ")
	w.WriteString(name)
	w.WriteByte('(')
	w.WriteString(args)
	w.WriteByte(')')

	if len(rets) > 0 {
		w.WriteByte(' ')
		w.WriteString(rets)
	}

	w.WriteString(" {\n")
}

func endMethod(w *codegen.File) {
	w.WriteString("}\n\n")
}

func genAssignFromDecoder(w *codegen.File, t string) {
	if c := t[0]; c < 'a' || c > 'z' {
		w.WriteString(".decode(d, v)\n")
		return
	}
	w.WriteString(" = d.decode")
	genCoderName(w, t)
	w.WriteString("()\n")
}

func genCoderName(w *codegen.File, t string) {
	switch t {
	case "bool", "boolean":
		w.WriteString("Bool")
	case "int8":
		w.WriteString("Int8")
	case "int16":
		w.WriteString("Int16")
	case "int32":
		w.WriteString("Int32")
	case "int64":
		w.WriteString("Int64")
	case "string":
		w.WriteString("String")
	case "bytes":
		w.WriteString("Bytes")
	default:
		panic("no coder for " + t)
	}
}

func genFieldDecl(w *codegen.File, f *schema.Field) {
	if desc := f.Desc; len(desc) != 0 {
		w.WriteString("// ")
		w.WriteString(desc)
		w.WriteByte('\n')
	}

	w.WriteString(f.Name)

	w.WriteByte(' ')

	if f.Type.Array {
		w.WriteString("[]")
	}
	elem := f.Type.Elem
	if elem == "bytes" {
		elem = "[]byte"
	}
	w.WriteString(elem)

	w.WriteString(" `json:")
	w.WriteQuoted(strcase.ToSnake(f.Name))
	w.WriteString("`\n")
}

func genFieldDecode(w *codegen.File, recv string, f *schema.Field) {
	v := f.Versions
	t := &f.Type

	begMethod(w, recv, "decode"+f.Name, "d *decoder, v int16", "")

	w.WriteString("if v < ")
	w.WriteInt(int64(v.Min), 10)
	if v.Max != -1 {
		w.WriteString(" || v > ")
		w.WriteInt(int64(v.Max), 10)
	}
	w.WriteString(" {\nreturn\n}\n")

	if t.Array {
		w.WriteString("a := make([]")
		w.WriteString(t.Elem)
		w.WriteString(", d.decodeInt32())\nfor i := range a {\na[i]")
		genAssignFromDecoder(w, t.Elem)
		w.WriteString("}\nm.")
		w.WriteString(f.Name)
		w.WriteString(" = a\n")
	} else {
		w.WriteString("m.")
		w.WriteString(f.Name)
		genAssignFromDecoder(w, t.Elem)
	}

	endMethod(w)
}

func genFieldDefault(w *codegen.File, f *schema.Field) {
	if f.Type.Array {
		w.WriteString("nil")
		return
	}
	switch t := f.Type.Elem; t {
	case "bool", "boolean":
		w.WriteBool(f.Default.Boolean())
	case "int8", "int16", "int32", "int64":
		i, err := strconv.ParseInt(t[3:], 10, 8)
		if err != nil {
			panic(err)
		}
		w.WriteInt(f.Default.Integer(int(i)), 10)
	case "string":
		w.WriteQuoted(f.Default.String())
	default:
		w.WriteString("nil")
	}
}

func genFieldEncode(w *codegen.File, recv string, f *schema.Field) {
	v := f.Versions
	t := &f.Type

	begMethod(w, recv, "encode"+f.Name, "e *encoder, v int16", "")

	w.WriteString("if v < ")
	w.WriteInt(int64(v.Min), 10)
	if v.Max != -1 {
		w.WriteString(" || v > ")
		w.WriteInt(int64(v.Max), 10)
	}
	w.WriteString(" {\nreturn\n}\n")

	if t.Array {
		w.WriteString("a := m.")
		w.WriteString(f.Name)
		w.WriteString("\nfor i := range a {\n")
		if c := t.Elem[0]; c < 'a' || c > 'z' {
			w.WriteString("a[i].encode(e, v)\n")
		} else {
			w.WriteString("e.encode")
			genCoderName(w, t.Elem)
			w.WriteString("(a[i])")
		}
		w.WriteString("}\n")
	} else {
		w.WriteString("e.encode")
		genCoderName(w, t.Elem)
		w.WriteString("(m.")
		w.WriteString(f.Name)
		w.WriteString(")\n")
	}

	endMethod(w)
}

func genMessage(w *codegen.File, m *schema.MessageData) {
	w.Write(m.Comments)
	genStructDecl(w, m, m.Name, m.Fields)

	for _, s := range m.CommonStructs {
		genStructDecl(w, m, s.Name, s.Fields)
	}
}

func genMessageMethods(w *codegen.File, m *schema.MessageData) {
	begMethod(w, m.Name, "isVersionFlexible", "v int16", "bool")
	w.WriteString("return ")
	genVersionCond(w, m.FlexibleVersions, m.ValidVersions)
	w.WriteByte('\n')
	endMethod(w)

	begMethod(w, m.Name, "isVersionValid", "v int16", "bool")
	w.WriteString("return ")
	genVersionCond(w, m.ValidVersions, nil)
	w.WriteByte('\n')
	endMethod(w)

	switch m.Type {
	case "header":
		// Nothing.
	case "request", "response":
		begMethod(w, m.Name, m.Type, "", "int16")
		w.WriteString("return ")
		w.WriteInt(int64(*m.ApiKey), 10)
		w.WriteByte('\n')
		endMethod(w)
	default:
		panic("unexpected message type: " + m.Type)
	}
}

func genStructDecl(w *codegen.File, m *schema.MessageData, name string, fields []*schema.Field) {
	w.WriteString("type ")
	w.WriteString(name)
	w.WriteString(" struct {\n")

	for i, f := range fields {
		if i != 0 {
			w.WriteByte('\n')
		}
		genFieldDecl(w, f)
	}

	w.WriteString("}\n\n")

	genStructReset(w, name, fields)
	genStructDecode(w, m, name, fields)
	genStructEncode(w, m, name, fields)

	if name == m.Name {
		genMessageMethods(w, m)
	}

	for _, f := range fields {
		if f.Fields == nil {
			continue
		}
		genStructDecl(w, m, f.Type.Elem, f.Fields)
	}
}

func genStructDecode(w *codegen.File, m *schema.MessageData, recv string, fields []*schema.Field) {
	begMethod(w, recv, "decode", "d *decoder, v int16", "")

	// if recv == m.Name {
	// 	w.WriteString("if !m.isVersionValid(v) {\npanic(errVersion)\n}\n")
	// }

	w.WriteString("m.Reset()\n")

	for _, f := range fields {
		w.WriteString("m.decode")
		w.WriteString(f.Name)
		w.WriteString("(d, v)\n")
	}

	endMethod(w)

	for _, f := range fields {
		genFieldDecode(w, recv, f)
	}
}

func genStructEncode(w *codegen.File, m *schema.MessageData, recv string, fields []*schema.Field) {
	begMethod(w, recv, "encode", "e *encoder, v int16", "")

	for _, f := range fields {
		w.WriteString("m.encode")
		w.WriteString(f.Name)
		w.WriteString("(e, v)\n")
	}

	endMethod(w)

	for _, f := range fields {
		genFieldEncode(w, recv, f)
	}
}

func genStructReset(w *codegen.File, recv string, fields []*schema.Field) {
	begMethod(w, recv, "Reset", "", "")

	for _, f := range fields {
		w.WriteString("m.")
		w.WriteString(f.Name)
		w.WriteString(" = ")
		genFieldDefault(w, f)
		w.WriteByte('\n')
	}

	endMethod(w)
}

func genVersionCond(w *codegen.File, v, p *schema.VersionRange) {
	if v == nil {
		if p == nil {
			panic("unset version range")
		}
		w.WriteString("false")
		return
	}
	if p != nil {
		if v.Min == -1 {
			if v.Max == -1 {
				w.WriteString("false")
				return
			}
			panic("unexpected version range")
		}
		if v.Max == -1 || v.Max == p.Max {
			if v.Min == p.Min {
				w.WriteString("true")
			} else {
				w.WriteString("v >= ")
				w.WriteInt(int64(v.Min), 10)
			}
			return
		}
		if v.Min == p.Min {
			w.WriteString("v <= ")
			w.WriteInt(int64(v.Max), 10)
			return
		}
	}
	w.WriteString("v >= ")
	w.WriteInt(int64(v.Min), 10)
	w.WriteString(" && v <= ")
	w.WriteInt(int64(v.Max), 10)
	return
}
