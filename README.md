# docker-machine-linode

Linode Driver Plugin for docker-machine.

## Install

First, docker-machine is required, documentation for how to install `docker-machine`
[is available here](https://docs.docker.com/machine/install-machine/).

Or, you can install `docker-machine` from source by running:

```bash
go get github.com/docker/machine
cd $GOPATH/src/github.com/docker/machine
make build
```

Then, install `docker-machine-linode` driver in the `$GOPATH` and add `$GOPATH/bin` to the `$PATH` environment variable.

```bash
go get github.com/displague/docker-machine-linode
cd $GOPATH/src/github.com/displague/docker-machine-linode
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
docker run --rm --it debian bash
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
| `linode-label` | `LINODE_LABEL` | *generated* | The Linode Instance `label`.  This `label` must be unique on the account.
| `linode-region` | `LINODE_REGION` | `us-east` | The Linode Instance `region` (see [here](https://api.linode.com/v4/regions))
| `linode-instance-type` | `LINODE_INSTANCE_TYPE` | `g6-standard-4` | The Linode Instance `type` (see [here](https://api.linode.com/v4/linode/types))
| `linode-image` | `LINODE_IMAGE` | `linode/ubuntu18.04` | The Linode Instance `image` which provides the Linux distribution (see [here](https://api.linode.com/v4/images)).
| `linode-kernel` | `LINODE_KERNEL` | `linode/grub2` | The Linux Instance `kernel` to boot.  `linode/grub2` will defer to the distribution kernel. (see [here](https://api.linode.com/v4/linode/kernels) (`?page=N`))
| `linode-ssh-port` | `LINODE_SSH_PORT` | `22` | The port that SSH is running on, needed for Docker Machine to provision the Linode.
| `linode-docker-port` | `LINODE_DOCKER_PORT` | `2376` | The TCP port of the Linode that Docker will be listening on
| `linode-swap-size` | `LINODE_SWAP_SIZE` | `512` | The amount of swap space provisioned on the Linode Instance


## Discussion / Help

Join us at [#linodego](https://gophers.slack.com/messages/CAG93EB2S) on the [gophers slack](https://gophers.slack.com)

## License

[MIT License](LICENSE)
