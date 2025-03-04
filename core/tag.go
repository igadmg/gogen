package core

import (
	"fmt"
	"go/ast"
	"maps"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

type TagData = map[string]any
type Tag struct {
	Data TagData
}

func MakeTag(tag string) (Tag, error) {
	m, err := UnmarshalTag(tag)
	if err != nil {
		return Tag{}, err
	}
	return Tag{Data: m}, nil
}

var errTag = fmt.Errorf("failed to parse tag")
var Tags = make([]string, 0)

func ParseTag(tag *ast.BasicLit) (t Tag, err error) {
	t = Tag{}
	if tag == nil {
		return t, errTag
	}

	vtag := strings.Trim(tag.Value, "`")
	stag := reflect.StructTag(vtag)

	for _, tagName := range Tags {
		gogtag, ok := stag.Lookup(tagName)
		if !ok {
			continue
		}

		var node TagData
		node, err = UnmarshalTag(gogtag)
		if err != nil {
			return
		}

		if t.Data == nil {
			t.Data = node
		} else {
			maps.Copy(t.Data, node)
		}
	}

	return
}

func UnmarshalTag(vtag string) (TagData, error) {
	m := TagData{}
	if err := yaml.Unmarshal([]byte("{"+vtag+"}"), m); err != nil {
		return nil, err
	}
	return m, nil
}

func (t Tag) IsEmpty() bool {
	return len(t.Data) == 0
}

func (t Tag) HasField(name string) bool {
	_, ok := t.Data[name]
	return ok
}

func (t Tag) GetField(name string) (string, bool) {
	v, ok := t.Data[name]
	if ok {
		switch vv := v.(type) {
		case string:
			return vv, true
		}
	}

	return "", false
}

func (t *Tag) SetField(name string, v any) {
	if t.Data == nil {
		t.Data = TagData{}
	}

	t.Data[name] = v
}

func (t Tag) GetObject(name string) (Tag, bool) {
	v, ok := t.Data[name]
	if ok {
		switch vv := v.(type) {
		case Tag:
			return vv, true
		case TagData:
			return Tag{vv}, true
		case yaml.Node:
			vm := TagData{}
			vv.Decode(vm)
			return Tag{vm}, true
		}
	}

	return Tag{}, false
}
