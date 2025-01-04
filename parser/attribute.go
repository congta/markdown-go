package parser

import (
	"bytes"
	"congta.com/qunmus/markdown/ast"
	"strings"
)

// attribute parses a (potential) block attribute and adds it to p.
func (p *Parser) attribute(data []byte) []byte {
	if len(data) < 3 {
		return data
	}
	i := 0
	if data[i] != '{' {
		return data
	}
	i++

	// last character must be a } otherwise it's not an attribute
	end := skipUntilChar(data, i, '\n')
	if data[end-1] != '}' {
		return data
	}
	// must have only one }
	if skipUntilChar(data, i, '}') != end-1 {
		return data
	}

	i = skipSpace(data, i)
	b := &ast.Attribute{Attrs: make(map[string][]byte)}

	esc := false
	quote := false
	trail := 0
Loop:
	for ; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\f', '\v':
			if quote {
				continue
			}
			chunk := data[trail+1 : i]
			if len(chunk) == 0 {
				trail = i
				continue
			}
			switch {
			case chunk[0] == '.':
				b.Classes = append(b.Classes, AutoConvertClass(chunk[1:]))
			case chunk[0] == '#':
				b.ID = chunk[1:]
			default:
				k, v := keyValue(chunk)
				if k != nil && v != nil {
					b.Attrs[string(k)] = v
				} else {
					// this is illegal in an attribute
					return data
				}
			}
			trail = i
		case '"':
			if esc {
				esc = !esc
				continue
			}
			quote = !quote
		case '\\':
			esc = !esc
		case '}':
			if esc {
				esc = !esc
				continue
			}
			chunk := data[trail+1 : i]
			if len(chunk) == 0 {
				return data
			}
			switch {
			case chunk[0] == '.':
				b.Classes = append(b.Classes, AutoConvertClass(chunk[1:]))
			case chunk[0] == '#':
				b.ID = chunk[1:]
			default:
				k, v := keyValue(chunk)
				if k != nil && v != nil {
					b.Attrs[string(k)] = v
				} else {
					return data
				}
			}
			i++
			break Loop
		default:
			esc = false
		}
	}

	p.attr = b
	return data[i:]
}

func AutoConvertClass(b []byte) []byte {
	class := AutoConvertStringClass(string(b))
	return []byte(class)
}

// AutoConvertStringClass not valid chars: " \ { } . #
func AutoConvertStringClass(class string) string {
	if strings.EqualFold(class, "1/1") {
		return "layui-col-xs12"
	} else if strings.EqualFold(class, "1/2") {
		return "layui-col-xs6"
	} else if strings.EqualFold(class, "1/3") {
		return "layui-col-xs4"
	} else if strings.EqualFold(class, "1/4") {
		return "layui-col-xs3"
	} else if strings.EqualFold(class, "1/6") {
		return "layui-col-xs2"
	} else if strings.EqualFold(class, "<->") { // todo(zhangfucheng) 待定
		return "flex-1"
	} else if strings.EqualFold(class, "!red") {
		return "ca-alert ca-alert-danger"
	} else if strings.EqualFold(class, "!yellow") {
		return "ca-alert ca-alert-warning"
	} else if strings.EqualFold(class, "!blue") {
		return "ca-alert ca-alert-primary"
	} else if strings.EqualFold(class, "!green") {
		return "ca-alert ca-alert-success"
	} else if strings.EqualFold(class, "!grey") || strings.EqualFold(class, "!gray") {
		return "ca-alert"
	}
	return class
}

// key="value" quotes are mandatory.
func keyValue(data []byte) ([]byte, []byte) {
	chunk := bytes.SplitN(data, []byte{'='}, 2)
	if len(chunk) != 2 {
		return nil, nil
	}
	key := chunk[0]
	value := chunk[1]

	if len(value) < 3 || len(key) == 0 {
		return nil, nil
	}
	if value[0] != '"' || value[len(value)-1] != '"' {
		return key, nil
	}
	return key, value[1 : len(value)-1]
}
