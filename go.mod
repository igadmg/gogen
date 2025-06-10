module github.com/igadmg/gogen

go 1.24.2

replace (
	deedles.dev/xiter => ../../pkg/xiter
	github.com/igadmg/goel => ../../cmd/goel
	github.com/igadmg/goex => ../../pkg/goex
)

require (
	deedles.dev/xiter v0.2.1
	github.com/igadmg/goecs v0.0.0-20250605212039-e73961886726
	github.com/igadmg/goex v0.0.0-20250511161240-125903d9a179
	github.com/stretchr/testify v1.10.0
	golang.org/x/tools v0.34.0
	gonum.org/v1/gonum v0.16.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
)
