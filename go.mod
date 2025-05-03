module github.com/igadmg/gogen

go 1.24.2

replace (
	deedles.dev/xiter => ../../pkg/xiter
	github.com/igadmg/goel => ../../cmd/goel
	github.com/igadmg/goex => ../../pkg/goex
)

require (
	deedles.dev/xiter v0.2.1
	github.com/igadmg/goex v0.0.0-20250502115452-bd40b01ba4eb
	github.com/stretchr/testify v1.10.0
	golang.org/x/tools v0.32.0
	gonum.org/v1/gonum v0.16.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/chewxy/math32 v1.11.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/igadmg/gamemath v0.0.0-20250502152201-14a551f600ad // indirect
	github.com/igadmg/goecs v0.0.0-20250503100802-ef527ce217f1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
)

