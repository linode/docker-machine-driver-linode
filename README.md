# docker-machine-linode

Linode Driver Plugin for docker-machine. **Requires docker-machine version > v.0.5.0-rc1**

## Install

First, docker-machine v0.5.0 rc2 is required, documentation for how to install `docker-machine`
[is available here](https://github.com/docker/machine/releases/tag/v0.5.0-rc2#Installation).

or you can install `docker-machine` from source code by running these commands

```bash
go get github.com/docker/machine
cd $GOPATH/src/github.com/docker/machine
make build
```

Then, install `docker-machine-linode` driver in the $GOPATH and add $GOPATH/bin to the $PATH env.

```bash
go get github.com/displague/docker-machine-linode
cd $GOPATH/src/github.com/displague/docker-machine-linode
make
make install
```

## Run

You will need a Linode APIv4 Personal Access Token.  Get one here: <https://developers.linode.com/api/v4#section/Personal-Access-Token>

```bash
docker-machine create -d linode --linode-token=<linode-token> --linode-root-pass=<linode-root-pass> linode
```

### Options

| Argument | Env | Default | Description
| --- | --- | --- | ---
| linode-token | LINODE_TOKEN | None | *REQUIRED* Linode APIv4 Token (see <https://developers.linode.com/api/v4#section/Personal-Access-Token>)
| linode-root-pass | LINODE_ROOT_PASSWORD | None | *REQUIRED* The Linode Instance `root_pass` (password assigned to the `root` account)
| linode-label | LINODE_LABEL | **generated** | The Linode Instance `label`.  This `label` must be unique on the account.
| linode-region | LINODE_REGION | `us-east` | The Linode Instance `region` (see <https://api.linode.com/v4/regions>)
| linode-instance-type | LINODE_INSTANCE_TYPE | `g6-standard-4` | The Linode Instance `type` (see <https://api.linode.com/v4/linode/types>)
| linode-image | LINODE_IMAGE | `linode/ubuntu18.04` | The Linode Instance `image` which provides the Linux distribution (see <https://api.linode.com/v4/images>).
| linode-kernel | LINODE_KERNEL | `linode/grub2` | The Linux Instance `kernel` to boot.  `linode/grub2` will defer to the distribution kernel. (see <https://api.linode.com/v4/linode/kernels> (`?page=N`))
| linode-ssh-port | LINODE_SSH_PORT | `22` | The port that SSH is running on, needed for Docker Machine to provision the Linode.
| linode-docker-port | LINODE_DOCKER_PORT | `2376` | The TCP port of the Linode that Docker will be listening on
| linode-swap-size | LINODE_SWAP_SIZE | `512` | The amount of swap space provisioned on the Linode Instance
