module github.com/igadmg/gogen

go 1.24

replace github.com/igadmg/goel => ../../cmd/goel

replace github.com/igadmg/goex => ../../pkg/goex

replace deedles.dev/xiter => ../../pkg/xiter

require (
	deedles.dev/xiter v0.1.1
	github.com/elliotchance/orderedmap/v3 v3.1.0
	github.com/igadmg/goex v0.0.0-20250226161117-f8fd602fe0c7
	github.com/stretchr/testify v1.10.0
	golang.org/x/tools v0.30.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
)
