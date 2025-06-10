package core

type TokenI interface {
	GetName() string
	GetFullName() string
	GetTag() Tag
	GetPackage() *Package
	SetPackage(pkg *Package)
}

type TokenM interface {
	SetTag(tag Tag)
}

type Token struct {
	Name    string   `hash:""`
	Tag     Tag      `hash:""`
	Package *Package `hash:""`
}

type TokenDto struct {
	Name    string  `yaml:""`
	Tag     TagData `yaml:""`
	Package string  `yaml:""`
}

func (t Token) MarshalYAML() (interface{}, error) {
	// Custom marshaling logic
	return TokenDto{
		Name:    t.Name,
		Tag:     t.Tag.Data,
		Package: t.Package.Name,
	}, nil
}

/*
func (t *Token) UnmarshalYAML(unmarshal func(interface{}) error) {
	// Custom unmarshaling logic
	type Alias Token
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := unmarshal(aux); err != nil {
		return err
	}
	return nil
}
*/

func MakeToken(name string, tag Tag) Token {
	return Token{
		Name: name,
		Tag:  tag,
	}
}

func (t Token) GetName() string {
	return t.Name
}

func (t Token) GetFullName() string {
	return t.Package.Name + "." + t.Name
}

func (t Token) GetTag() Tag {
	return t.Tag
}

func (t *Token) SetTag(tag Tag) {
	t.Tag = tag
}

func (t Token) GetPackage() *Package {
	return t.Package
}

func (t *Token) SetPackage(pkg *Package) {
	t.Package = pkg
}
