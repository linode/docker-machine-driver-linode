# docker-machine-driver-linode

[![GoDoc](https://godoc.org/github.com/linode/docker-machine-driver-linode?status.svg)](https://godoc.org/github.com/linode/docker-machine-driver-linode)
[![Go Report Card](https://goreportcard.com/badge/github.com/linode/docker-machine-driver-linode)](https://goreportcard.com/report/github.com/linode/docker-machine-driver-linode)
[![CircleCI](https://circleci.com/gh/linode/docker-machine-driver-linode.svg?style=svg)](https://circleci.com/gh/linode/docker-machine-driver-linode)
[![GitHub release](https://img.shields.io/github/release/linode/docker-machine-driver-linode.svg)](https://github.com/linode/docker-machine-driver-linode/releases/)

Linode Driver Plugin for docker-machine.

## Install

`docker-machine` is required, [see the installation documentation](https://docs.docker.com/machine/install-machine/).

Then, install the latest release for your environment from the [releases list](https://github.com/linode/docker-machine-driver-linode/releases).

### Installing from source

If you would rather build from source, you will need to have a working `go` 1.11+ environment,

```bash
eval $(go env)
export PATH="$PATH:$GOPATH/bin"
```

You can then install `docker-machine` from source by running:

```bash
go get github.com/docker/machine
cd $GOPATH/src/github.com/docker/machine
make build
```

And then compile the `docker-machine-driver-linode` driver:

```bash
go get github.com/linode/docker-machine-driver-linode
cd $GOPATH/src/github.com/linode/docker-machine-driver-linode
make install
```

## Run

You will need a Linode APIv4 Personal Access Token.  Get one here: <https://developers.linode.com/api/v4#section/Personal-Access-Token>

```bash
docker-machine create -d linode --linode-token=<linode-token> --linode-root-pass=<linode-root-pass> linode
```

### Full Example

```bash
LINODE_TOKEN=e332cf8e1a78427f1368a5a0a67946ad1e7c8e28e332cf8e1a78427f1368a5a0 # Should be 65 lowercase hex chars
LINODE_ROOT_PASSWORD=$(openssl rand -base64 32); echo Password for root: $LINODE_ROOT_PASSWORD

docker-machine create -d linode --linode-token=$LINODE_TOKEN --linode-root-pass=$LINODE_ROOT_PASSWORD linode
eval $(docker-machine env linode)
docker run --rm -it debian bash
```

```bash
$ docker-machine ls
NAME      ACTIVE   DRIVER   STATE     URL                        SWARM   DOCKER        ERRORS
linode    *        linode   Running   tcp://45.79.139.196:2376           v18.05.0-ce

$ docker-machine rm linode
About to remove linode
WARNING: This action will delete both local reference and remote instance.
Are you sure? (y/n): y
(default) Removing linode: 8753395
Successfully removed linode
```

### Options

| Argument | Env | Default | Description
| --- | --- | --- | ---
| `linode-token` | `LINODE_TOKEN` | None | **required** Linode APIv4 Token (see [here](https://developers.linode.com/api/v4#section/Personal-Access-Token))
| `linode-root-pass` | `LINODE_ROOT_PASSWORD` | None | **required** The Linode Instance `root_pass` (password assigned to the `root` account)
| `linode-label` | `LINODE_LABEL` | *generated* | The Linode Instance `label`, unless overridden this will match the docker-machine name.  This `label` must be unique on the account.
| `linode-region` | `LINODE_REGION` | `us-east` | The Linode Instance `region` (see [here](https://api.linode.com/v4/regions))
| `linode-instance-type` | `LINODE_INSTANCE_TYPE` | `g6-standard-4` | The Linode Instance `type` (see [here](https://api.linode.com/v4/linode/types))
| `linode-image` | `LINODE_IMAGE` | `linode/ubuntu18.04` | The Linode Instance `image` which provides the Linux distribution (see [here](https://api.linode.com/v4/images)).
| `linode-kernel` | `LINODE_KERNEL` | `linode/grub2` | The Linux Instance `kernel` to boot.  `linode/grub2` will defer to the distribution kernel. (see [here](https://api.linode.com/v4/linode/kernels) (`?page=N`))
| `linode-ssh-port` | `LINODE_SSH_PORT` | `22` | The port that SSH is running on, needed for Docker Machine to provision the Linode.
| `linode-ssh-user` | `LINODE_SSH_USER` | `root` | The user as which docker-machine should log in to the Linode instance to install Docker.  This user must have passwordless sudo.
| `linode-docker-port` | `LINODE_DOCKER_PORT` | `2376` | The TCP port of the Linode that Docker will be listening on
| `linode-swap-size` | `LINODE_SWAP_SIZE` | `512` | The amount of swap space provisioned on the Linode Instance
| `linode-stackscript` | `LINODE_STACKSCRIPT` | None | Specifies the Linode StackScript to use to create the instance, either by numeric ID, or using the form *username*/*label*.
| `linode-stackscript-data` | `LINODE_STACKSCRIPT_DATA` | None | A JSON string specifying data that is passed (via UDF) to the selected StackScript.
| `linode-create-private-ip` | `LINODE_CREATE_PRIVATE_IP` | None | A flag specifying to create private IP for the Linode instance.

## Debugging

Detailed run output will be emitted when using the LinodeGo `LINODE_DEBUG=1` option along with the `docker-machine` `--debug` option.

```bash
LINODE_DEBUG=1 docker-machine --debug  create -d linode --linode-token=$LINODE_TOKEN --linode-root-pass=$ROOT_PASS machinename
```

## Discussion / Help

Join us at [#linodego](https://gophers.slack.com/messages/CAG93EB2S) on the [gophers slack](https://gophers.slack.com)

## License

[MIT License](LICENSE)
