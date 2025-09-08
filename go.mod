module github.com/linode/docker-machine-driver-linode

go 1.24.0

toolchain go1.24.1

// This replacement is necessary to support Docker versions > v20.x.x
// which provide critical security fixes.
replace github.com/docker/machine => gitlab.com/gitlab-org/ci-cd/docker-machine v0.16.2-gitlab.27

require (
	github.com/docker/machine v0.16.2
	github.com/google/go-cmp v0.7.0
	github.com/linode/linodego v1.56.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/oauth2 v0.31.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-resty/resty/v2 v2.16.5 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
