package main

import (
	linode "github.com/displague/docker-machine-linode"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(linode.NewDriver("", ""))
}
