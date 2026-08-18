module github.com/linode/docker-machine-driver-linode

go 1.25.8

// This replacement is necessary to support Docker versions > v20.x.x
// which provide critical security fixes.
replace github.com/docker/machine => gitlab.com/gitlab-org/ci-cd/docker-machine v0.16.2-gitlab.46

require (
	github.com/docker/machine v0.16.2
	github.com/google/go-cmp v0.7.0
	github.com/linode/linodego/v2 v2.5.0
	github.com/stretchr/testify v1.12.0
	golang.org/x/oauth2 v0.36.0
)

require (
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
