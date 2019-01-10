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

### Options

| Argument | Env | Default | Description
| --- | --- | --- | ---
| `linode-token` | `LINODE_TOKEN` | None | **required** Linode APIv4 Token (see [here](https://developers.linode.com/api/v4#section/Personal-Access-Token))
| `linode-root-pass` | `LINODE_ROOT_PASSWORD` | *generated* | The Linode Instance `root_pass` (password assigned to the `root` account)
| `linode-authorized-users` | `LINODE_AUTHORIZED_USERS` | None | Linode user accounts (separated by commas) whose Linode SSH keys will be permitted root access to the created node
| `linode-label` | `LINODE_LABEL` | *generated* | The Linode Instance `label`, unless overridden this will match the docker-machine name.  This `label` must be unique on the account.
| `linode-region` | `LINODE_REGION` | `us-east` | The Linode Instance `region` (see [here](https://api.linode.com/v4/regions))
| `linode-instance-type` | `LINODE_INSTANCE_TYPE` | `g6-standard-4` | The Linode Instance `type` (see [here](https://api.linode.com/v4/linode/types))
| `linode-image` | `LINODE_IMAGE` | `linode/ubuntu18.04` | The Linode Instance `image` which provides the Linux distribution (see [here](https://api.linode.com/v4/images)).
| `linode-ssh-port` | `LINODE_SSH_PORT` | `22` | The port that SSH is running on, needed for Docker Machine to provision the Linode.
| `linode-ssh-user` | `LINODE_SSH_USER` | `root` | The user as which docker-machine should log in to the Linode instance to install Docker.  This user must have passwordless sudo.
| `linode-docker-port` | `LINODE_DOCKER_PORT` | `2376` | The TCP port of the Linode that Docker will be listening on
| `linode-swap-size` | `LINODE_SWAP_SIZE` | `512` | The amount of swap space provisioned on the Linode Instance
| `linode-stackscript` | `LINODE_STACKSCRIPT` | None | Specifies the Linode StackScript to use to create the instance, either by numeric ID, or using the form *username*/*label*.
| `linode-stackscript-data` | `LINODE_STACKSCRIPT_DATA` | None | A JSON string specifying data that is passed (via UDF) to the selected StackScript.
| `linode-create-private-ip` | `LINODE_CREATE_PRIVATE_IP` | None | A flag specifying to create private IP for the Linode instance.
| `linode-ua-prefix` | `LINODE_UA_PREFIX` | None | Prefix the User-Agent in Linode API calls with some 'product/version'

## Notes

* When using the `linode/containerlinux` `linode-image`, the `linode-ssh-user` will default to `core`
* A `linode-root-pass` will be generated if not provided.  This password will not be shown. Rely on `docker-machine ssh` or [Linode's Rescue features](https://www.linode.com/docs/quick-answers/linode-platform/reset-the-root-password-on-your-linode/) to access the node directly.

## Debugging

Detailed run output will be emitted when using the LinodeGo `LINODE_DEBUG=1` option along with the `docker-machine` `--debug` option.

```bash
LINODE_DEBUG=1 docker-machine --debug  create -d linode --linode-token=$LINODE_TOKEN machinename
```

## Examples

### Simple Example

```bash
LINODE_TOKEN=e332cf8e1a78427f1368a5a0a67946ad1e7c8e28e332cf8e1a78427f1368a5a0 # Should be 65 lowercase hex chars

docker-machine create -d linode --linode-token=$LINODE_TOKEN linode
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

### Provisioning Docker Swarm

The following script serves as an example for creating a [Docker Swarm](https://docs.docker.com/engine/swarm/) with master and worker nodes using the Linode Docker machine driver and private networking.

This script is provided for demonstrative use.  A production swarm environment would require hardening.

1. Create an `install.sh` bash script using the source below.  Run `bash install.sh` and provide a Linode APIv4 Token when prompted.

    ```sh
    #!/bin/bash
    set -e

    read -p "Linode Token: " LINODE_TOKEN
    # LINODE_TOKEN=...
    LINODE_ROOT_PASSWORD=$(openssl rand -base64 32); echo Password for root: $LINODE_ROOT_PASSWORD
    LINODE_REGION=eu-central

    create_node() {
        local name=$1
        docker-machine create \
        -d linode \
        --linode-label=$name \
        --linode-instance-type=g6-nanode-1 \
        --linode-image=linode/ubuntu18.04 \
        --linode-region=$LINODE_REGION \
        --linode-token=$LINODE_TOKEN \
        --linode-root-pass=$LINODE_ROOT_PASSWORD \
        --linode-create-private-ip \
        $name
    }

    get_private_ip() {
        local name=$1
        docker-machine inspect  -f '{{.Driver.PrivateIPAddress}}' $name
    }

    init_swarm_master() {
        local name=$1
        local ip=$(get_private_ip $name)
        docker-machine ssh $name "docker swarm init --advertise-addr ${ip}"
    }

    init_swarm_worker() {
        local master_name=$1
        local worker_name=$2
        local master_addr=$(get_private_ip $master_name):2377
        local join_token=$(docker-machine ssh $master_name "docker swarm join-token worker -q")
        docker-machine ssh $worker_name "docker swarm join --token=${join_token} ${master_addr}"
    }

    # create master node
    create_node master01

    # create worker node
    create_node worker01

    # init swarm master
    init_swarm_master master01

    # init swarm worker
    init_swarm_worker master01 worker01
    ```

1. After provisioning succeeds, check the Docker Swarm status.  The output should show active an swarm leader and worker.

    ```sh
    $ eval $(docker-machine env master01)
    $ docker node ls

    ID                            HOSTNAME            STATUS              AVAILABILITY        MANAGER STATUS      ENGINE VERSION
    f8x7zutegt2dn1imeiw56v9hc *   master01            Ready               Active              Leader              18.09.0
    ja8b3ut6uaivz5hf98gah469y     worker01            Ready               Active                                  18.09.0
    ```

1. [Create and scale Docker services](https://docs.docker.com/engine/reference/commandline/service_create/) (left as an excercise for the reader).

    ```bash
    $ docker service create --name my-service --replicas 3 nginx:alpine
    $ docker node ps master01 worker01
    ID                  NAME                IMAGE               NODE                DESIRED STATE       CURRENT STATE           ERROR               PORTS
    7cggbrqfqopn         \_ my-service.1    nginx:alpine        master01            Running             Running 4 minutes ago
    7cggbrqfqopn         \_ my-service.1    nginx:alpine        master01            Running             Running 4 minutes ago
    v7c1ni5q43uu        my-service.2        nginx:alpine        worker01            Running             Running 4 minutes ago
    2w6d8o3hdyh4        my-service.3        nginx:alpine        worker01            Running             Running 4 minutes ago
    ```

1. Cleanup the resources

    ```sh
    docker-machine rm worker01 -y
    docker-machine rm master01 -y
    ```

## Discussion / Help

Join us at [#linodego](https://gophers.slack.com/messages/CAG93EB2S) on the [gophers slack](https://gophers.slack.com)

## License

[MIT License](LICENSE)
