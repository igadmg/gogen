module github.com/igadmg/gogen

go 1.24

replace (
	deedles.dev/xiter => ../../pkg/xiter
	github.com/igadmg/goel => ../../cmd/goel
	github.com/igadmg/goex => ../../pkg/goex
)

require (
	deedles.dev/xiter v0.2.1
	github.com/igadmg/goex v0.0.0-20250407220752-712c023573b8
	github.com/stretchr/testify v1.10.0
	golang.org/x/tools v0.32.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
)
