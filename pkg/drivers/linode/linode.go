package linode

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

// Driver is the implementation of BaseDriver interface
type Driver struct {
	*drivers.BaseDriver
	client *linodego.Client

	APIToken   string
	IPAddress  string
	DockerPort int

	InstanceID    int
	InstanceLabel string

	Region         string
	InstanceType   string
	RootPassword   string
	SSHPort        int
	InstanceImage  string
	InstanceKernel string
	SwapSize       int
}

const (
	// VERSION represents the semver version of the package
	VERSION               = "0.0.9"
	defaultSSHPort        = 22
	defaultSSHUser        = "root"
	defaultInstanceImage  = "linode/ubuntu18.04"
	defaultRegion         = "us-east"
	defaultInstanceType   = "g6-standard-4"
	defaultInstanceKernel = "linode/grub2"
	defaultSwapSize       = 512
	defaultDockerPort     = 2376
)

// NewDriver
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		InstanceImage: defaultInstanceImage,
		InstanceType:  defaultInstanceType,
		Region:        defaultRegion,
		SwapSize:      defaultSwapSize,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// Get Linode Client
func (d *Driver) getClient() *linodego.Client {
	if d.client == nil {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: d.APIToken})

		oauth2Client := &http.Client{
			Transport: &oauth2.Transport{
				Source: tokenSource,
			},
		}

		client := linodego.NewClient(oauth2Client)
		client.SetUserAgent(fmt.Sprintf("docker-machine-driver-%s/v%s linodego/%s", d.DriverName(), VERSION, linodego.Version))
		client.SetDebug(true)
		d.client = &client
	}
	return d.client
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "linode"
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
// Note that currently the IP Address is cached
func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "LINODE_TOKEN",
			Name:   "linode-token",
			Usage:  "Linode API Token",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_ROOT_PASSWORD",
			Name:   "linode-root-pass",
			Usage:  "Root Password",
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_LABEL",
			Name:   "linode-label",
			Usage:  "Linode Instance Label",
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_REGION",
			Name:   "linode-region",
			Usage:  "Specifies the region (location) of the Linode instance",
			Value:  defaultRegion, // "us-central", "ap-south", "eu-central", ...
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_INSTANCE_TYPE",
			Name:   "linode-instance-type",
			Usage:  "Specifies the Linode Instance type which determines CPU, memory, disk size, etc.",
			Value:  defaultInstanceType, // "g6-nanode-1", g6-highmem-2, ...
		},
		mcnflag.IntFlag{
			EnvVar: "LINODE_SSH_PORT",
			Name:   "linode-ssh-port",
			Usage:  "Linode Instance SSH Port",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_IMAGE",
			Name:   "linode-image",
			Usage:  "Specifies the Linode Instance image which determines the OS distribution and base files",
			Value:  defaultInstanceImage, // "linode/ubuntu18.04", "linode/arch", ...
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_KERNEL",
			Name:   "linode-kernel",
			Usage:  "Linode Instance Kernel",
			Value:  defaultInstanceKernel, // linode/latest-64bit, ..
		},
		mcnflag.IntFlag{
			EnvVar: "LINODE_DOCKER_PORT",
			Name:   "linode-docker-port",
			Usage:  "Docker Port",
			Value:  defaultDockerPort,
		},
		mcnflag.IntFlag{
			EnvVar: "LINODE_SWAP_SIZE",
			Name:   "linode-swap-size",
			Usage:  "Linode Instance Swap Size (MB)",
			Value:  defaultSwapSize,
		},
	}
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}

	return d.SSHUser
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIToken = flags.String("linode-token")
	d.Region = flags.String("linode-region")
	d.InstanceType = flags.String("linode-instance-type")
	d.RootPassword = flags.String("linode-root-pass")
	d.SSHPort = flags.Int("linode-ssh-port")
	d.InstanceImage = flags.String("linode-image")
	d.InstanceKernel = flags.String("linode-kernel")
	d.InstanceLabel = flags.String("linode-label")
	d.SwapSize = flags.Int("linode-swap-size")
	d.DockerPort = flags.Int("linode-docker-port")

	d.SetSwarmConfigFromFlags(flags)

	if d.APIToken == "" {
		return fmt.Errorf("linode driver requires the --linode-token option")
	}

	if d.RootPassword == "" {
		return fmt.Errorf("linode driver requires the --linode-root-pass option")
	}

	if len(d.InstanceLabel) == 0 {
		d.InstanceLabel = d.GetMachineName()
	}
	if strings.Contains(d.InstanceImage, "linode/containerlinux") {
		d.SSHUser = "core"
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	// TODO linode-stackscript-file should be read and uploaded (private), then used for boot.
	// RevNote could be sha256 of file so the file can be referenced instead of reuploaded.
	// linode-stackscript would let the user specify an existing id
	// linode-stackscript-data would need to be a json input
	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Info("Creating Linode machine instance...")

	publicKey, err := d.createSSHKey()
	if err != nil {
		return err
	}

	client := d.getClient()

	// Create a linode
	log.Info("Creating linode instance")
	createOpts := linodego.InstanceCreateOptions{
		Region:         d.Region,
		Type:           d.InstanceType,
		Label:          d.InstanceLabel,
		RootPass:       d.RootPassword,
		AuthorizedKeys: []string{strings.TrimSpace(publicKey)},
		Image:          d.InstanceImage,
		SwapSize:       &d.SwapSize,
	}

	linode, err := client.CreateInstance(context.TODO(), createOpts)
	if err != nil {
		return err
	}

	d.InstanceID = linode.ID
	d.InstanceLabel = linode.Label

	for _, address := range linode.IPv4 {
		if private := privateIP(*address); !private {
			d.IPAddress = address.String()
			break
		}
	}

	if d.IPAddress == "" {
		return errors.New("Linode IP Address is not found")
	}

	log.Debugf("Created Linode Instance %s (%d), IP address %s",
		d.InstanceLabel,
		d.InstanceID,
		d.IPAddress)

	if err != nil {
		return err
	}

	log.Info("Waiting for Machine Running...")
	if _, err := client.WaitForInstanceStatus(context.TODO(), d.InstanceID, linodego.InstanceRunning, 180); err != nil {
		return fmt.Errorf("wait for machine running failed: %s", err)
	}

	return nil
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:%d", ip, d.DockerPort), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	linode, err := d.getClient().GetInstance(context.TODO(), d.InstanceID)
	if err != nil {
		return state.Error, err
	}

	switch linode.Status {
	case linodego.InstanceRunning:
		return state.Running, nil
	case linodego.InstanceOffline,
		linodego.InstanceRebuilding,
		linodego.InstanceMigrating:
		return state.Stopped, nil
	case linodego.InstanceShuttingDown, linodego.InstanceDeleting:
		return state.Stopping, nil
	case linodego.InstanceProvisioning,
		linodego.InstanceRebooting,
		linodego.InstanceBooting,
		linodego.InstanceCloning,
		linodego.InstanceRestoring:
		return state.Starting, nil

	}

	// deleting, migrating, rebuilding, cloning, restoring ...
	return state.None, nil
}

// Start a host
func (d *Driver) Start() error {
	log.Debug("Start...")
	err := d.getClient().BootInstance(context.TODO(), d.InstanceID, 0)
	return err
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	log.Debug("Stop...")
	err := d.getClient().ShutdownInstance(context.TODO(), d.InstanceID)
	return err
}

// Remove a host
func (d *Driver) Remove() error {
	client := d.getClient()
	log.Infof("Removing linode: %d", d.InstanceID)
	if err := client.DeleteInstance(context.TODO(), d.InstanceID); err != nil {
		if apiErr, ok := err.(*linodego.Error); ok && apiErr.Code == 404 {
			log.Debug("Linode was already removed")
			return nil
		}

		return err
	}
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	log.Debug("Restarting...")
	err := d.getClient().RebootInstance(context.TODO(), d.InstanceID, 0)
	return err
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	log.Debug("Killing...")
	err := d.getClient().ShutdownInstance(context.TODO(), d.InstanceID)
	return err
}

func (d *Driver) createSSHKey() (string, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

// publicSSHKeyPath is always SSH Key Path appended with ".pub"
func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// privateIP determines if an IP is for private use (RFC1918)
// https://stackoverflow.com/a/41273687
func privateIP(ip net.IP) bool {
	private := false
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
	private = private24BitBlock.Contains(ip) || private20BitBlock.Contains(ip) || private16BitBlock.Contains(ip)
	return private
}
