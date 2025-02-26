module github.com/linode/docker-machine-driver-linode

go 1.22
toolchain go1.23.6

// This replacement is necessary to support Docker versions > v20.x.x
// which provide critical security fixes.
replace github.com/docker/machine => gitlab.com/gitlab-org/ci-cd/docker-machine v0.16.2-gitlab.27

require (
	github.com/docker/machine v0.16.2
	github.com/google/go-cmp v0.6.0
	github.com/linode/linodego v1.47.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/oauth2 v0.27.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-resty/resty/v2 v2.16.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
