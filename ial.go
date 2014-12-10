// IAL implements

package mmark

import "bytes"
import (
	"sort"
	"strings"
)

// One or more of these can be attached to block elements

type IAL struct {
	id    string            // #id
	class map[string]bool   // 0 or more .class
	attr  map[string]string // key=value pairs
}

func newIAL() *IAL {
	return &IAL{class: make(map[string]bool), attr: make(map[string]string)}
}

// Parsing and thus detecting an IAL. Return a valid *IAL or nil.
// IAL can have #id, .class or key=value element seperated by spaces, that may be escaped
func (p *parser) isIAL(data []byte) int {
	esc := false
	quote := false
	ialB := 0
	ial := newIAL()
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ':
			if quote {
				continue
			}
			chunk := data[ialB+1 : i]
			if len(chunk) == 0 {
				ialB = i
				continue
			}
			switch {
			case chunk[0] == '.':
				ial.class[string(chunk[1:])] = true
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				}
			}
			ialB = i
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
			// if this is mainmatter, frontmatter, or backmatter it isn't an IAL.
			s := string(data[1:i])
			switch s {
			case "frontmatter":
				fallthrough
			case "mainmatter":
				fallthrough
			case "backmatter":
				return 0
			}
			chunk := data[ialB+1 : i]
			if len(chunk) == 0 {
				return i + 1
			}
			switch {
			case chunk[0] == '.':
				ial.class[string(chunk[1:])] = true
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				}
			}
			p.ial = p.ial.add(ial)
			return i + 1
		default:
			esc = false
		}
	}
	return 0
}

func parseKeyValue(chunk []byte) (string, string) {
	chunks := bytes.SplitN(chunk, []byte{'='}, 2)
	if len(chunks) != 2 {
		return "", ""
	}
	chunks[1] = bytes.Replace(chunks[1], []byte{'"'}, nil, -1)
	return string(chunks[0]), string(chunks[1])
}

// Add IAL to another, overwriting the #id, collapsing classes and attributes
func (i *IAL) add(j *IAL) *IAL {
	if i == nil {
		return j
	}
	if j.id != "" {
		i.id = j.id
	}
	for k, c := range j.class {
		i.class[k] = c
	}
	for k, a := range j.attr {
		i.attr[k] = a
	}
	return i
}

// String renders an IAL and returns a string that can be included in the tag:
// class="class" anchor="id" key="value". The string s has a space as the first character.k
func (i *IAL) String() (s string) {
	if i == nil {
		return ""
	}

	// some fluff needed to make this all sorted.
	if i.id != "" {
		s = " anchor=\"" + i.id + "\""
	}

	keys := make([]string, 0, len(i.class))
	for k, _ := range i.class {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		s += " class=\"" + strings.Join(keys, " ") + "\""
	}

	keys = keys[:0]
	for k, _ := range i.attr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attr := make([]string, len(keys))
	for j, k := range keys {
		v := i.attr[k]
		attr[j] = k + "=\"" + v + "\""
	}
	if len(keys) > 0 {
		s += " " + strings.Join(attr, " ")
	}
	return s
}

// GetOrDefaultAttr return the value under key (and delete the key from the attributes) or
// returns the default value if the key is not found.
func (i *IAL) GetOrDefaultAttr(key, def string) string {
	v := i.attr[key]
	delete(i.attr, key)
	if v != "" {
		return v
	}
	return def
}

// GetOrDefaultId return the Id or
// returns the default value if the id not empty.
func (i *IAL) GetOrDefaultId(id string) string {
	if i.id != "" {
		j := i.id
		i.id = ""
		return j
	}
	return id
}
