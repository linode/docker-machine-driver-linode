package main

import (
	"github.com/displague/docker-machine-driver-linode/pkg/drivers/linode"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(linode.NewDriver("", ""))
}
