package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/linode/docker-machine-driver-linode/pkg/drivers/linode"
)

func main() {
	plugin.RegisterDriver(linode.NewDriver("", ""))
}
