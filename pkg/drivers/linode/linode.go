package linode

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
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

	APIToken         string
	UserAgentPrefix  string
	IPAddress        string
	PrivateIPAddress string
	CreatePrivateIP  bool
	DockerPort       int

	InstanceID    int
	InstanceLabel string

	Region          string
	InstanceType    string
	RootPassword    string
	AuthorizedUsers string
	SSHPort         int
	InstanceImage   string
	SwapSize        int

	StackScriptID    int
	StackScriptUser  string
	StackScriptLabel string
	StackScriptData  map[string]string

	Tags string
}

var (
	// VERSION represents the semver version of the package
	VERSION = "devel"
)

const (
	defaultSSHPort       = 22
	defaultSSHUser       = "root"
	defaultInstanceImage = "linode/ubuntu18.04"
	defaultRegion        = "us-east"
	defaultInstanceType  = "g6-standard-4"
	defaultSwapSize      = 512
	defaultDockerPort    = 2376

	defaultContainerLinuxSSHUser = "core"
)

// NewDriver creates and returns a new instance of the Linode driver
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

// getClient prepares the Linode APIv4 Client
func (d *Driver) getClient() *linodego.Client {
	if d.client == nil {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: d.APIToken})

		oauth2Client := &http.Client{
			Transport: &oauth2.Transport{
				Source: tokenSource,
			},
		}

		ua := fmt.Sprintf("docker-machine-driver-%s/%s linodego/%s", d.DriverName(), VERSION, linodego.Version)

		client := linodego.NewClient(oauth2Client)
		if len(d.UserAgentPrefix) > 0 {
			ua = fmt.Sprintf("%s %s", d.UserAgentPrefix, ua)
		}

		client.SetUserAgent(ua)
		client.SetDebug(true)
		d.client = &client
	}
	return d.client
}

func createRandomRootPassword() (string, error) {
	rawRootPass := make([]byte, 50)
	_, err := rand.Read(rawRootPass)
	if err != nil {
		return "", fmt.Errorf("Failed to generate random password")
	}
	rootPass := base64.StdEncoding.EncodeToString(rawRootPass)
	return rootPass, nil
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
			EnvVar: "LINODE_AUTHORIZED_USERS",
			Name:   "linode-authorized-users",
			Usage:  "Linode user accounts (seperated by commas) whose Linode SSH keys will be permitted root access to the created node",
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
			EnvVar: "LINODE_SSH_USER",
			Name:   "linode-ssh-user",
			Usage:  "Specifies the user as which docker-machine should log in to the Linode instance to install Docker.",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_IMAGE",
			Name:   "linode-image",
			Usage:  "Specifies the Linode Instance image which determines the OS distribution and base files",
			Value:  defaultInstanceImage, // "linode/ubuntu18.04", "linode/arch", ...
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
		mcnflag.StringFlag{
			EnvVar: "LINODE_STACKSCRIPT",
			Name:   "linode-stackscript",
			Usage:  "Specifies the Linode StackScript to use to create the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_STACKSCRIPT_DATA",
			Name:   "linode-stackscript-data",
			Usage:  "A JSON string specifying data for the selected StackScript",
			Value:  "",
		},
		mcnflag.BoolFlag{
			EnvVar: "LINODE_CREATE_PRIVATE_IP",
			Name:   "linode-create-private-ip",
			Usage:  "Create private IP for the instance",
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_UA_PREFIX",
			Name:   "linode-ua-prefix",
			Usage:  fmt.Sprintf("Prefix the User-Agent in Linode API calls with some 'product/version'"),
		},
		mcnflag.StringFlag{
			EnvVar: "LINODE_TAGS",
			Name:   "linode-tags",
			Usage:  fmt.Sprintf("A comma separated list of tags to apply to the the Linode resource"),
		},
	}
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = defaultSSHPort
	}

	return d.SSHPort, nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		if strings.Contains(d.InstanceImage, "linode/containerlinux") {
			d.SSHUser = defaultContainerLinuxSSHUser
		} else {
			d.SSHUser = defaultSSHUser
		}
	}

	return d.SSHUser
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIToken = flags.String("linode-token")
	d.Region = flags.String("linode-region")
	d.InstanceType = flags.String("linode-instance-type")
	d.AuthorizedUsers = flags.String("linode-authorized-users")
	d.RootPassword = flags.String("linode-root-pass")
	d.SSHPort = flags.Int("linode-ssh-port")
	d.SSHUser = flags.String("linode-ssh-user")
	d.InstanceImage = flags.String("linode-image")
	d.InstanceLabel = flags.String("linode-label")
	d.SwapSize = flags.Int("linode-swap-size")
	d.DockerPort = flags.Int("linode-docker-port")
	d.CreatePrivateIP = flags.Bool("linode-create-private-ip")
	d.UserAgentPrefix = flags.String("linode-ua-prefix")
	d.Tags = flags.String("linode-tags")

	d.SetSwarmConfigFromFlags(flags)

	if d.APIToken == "" {
		return fmt.Errorf("linode driver requires the --linode-token option")
	}

	stackScript := flags.String("linode-stackscript")
	if stackScript != "" {
		sid, err := strconv.Atoi(stackScript)
		if err == nil {
			d.StackScriptID = sid
		} else {
			ss := strings.SplitN(stackScript, "/", 2)
			if len(ss) != 2 {
				return fmt.Errorf("linode StackScripts must be specified using username/label syntax, or using their identifier")
			}

			d.StackScriptUser = ss[0]
			d.StackScriptLabel = ss[1]
		}

		stackScriptDataStr := flags.String("linode-stackscript-data")
		if stackScriptDataStr != "" {
			err := json.Unmarshal([]byte(stackScriptDataStr), &d.StackScriptData)
			if err != nil {
				return fmt.Errorf("linode StackScript data must be valid JSON: %s", err)
			}
		}
	}

	if len(d.InstanceLabel) == 0 {
		d.InstanceLabel = d.GetMachineName()
	}

	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	// TODO(displague) linode-stackscript-file should be read and uploaded (private), then used for boot.
	// RevNote could be sha256 of file so the file can be referenced instead of reuploaded.

	client := d.getClient()

	if d.RootPassword == "" {
		log.Info("Generating a secure disposable linode-root-pass...")
		var err error
		d.RootPassword, err = createRandomRootPassword()
		if err != nil {
			return err
		}
	}

	if d.StackScriptUser != "" {
		/* N.B. username isn't on the list of filterable fields, however
		   adding it doesn't make anything fail, so if it becomes
		   filterable in future this will become more efficient */
		options := map[string]string{
			"username": d.StackScriptUser,
			"label":    d.StackScriptLabel,
		}
		b, err := json.Marshal(options)
		if err != nil {
			return err
		}
		opts := linodego.NewListOptions(0, string(b))
		stackscripts, err := client.ListStackscripts(context.TODO(), opts)
		if err != nil {
			return err
		}
		var script *linodego.Stackscript
		for _, s := range stackscripts {
			if s.Username == d.StackScriptUser {
				script = &s
				break
			}
		}
		if script == nil {
			return fmt.Errorf("StackScript not found: %s/%s", d.StackScriptUser, d.StackScriptLabel)
		}

		d.StackScriptUser = script.Username
		d.StackScriptLabel = script.Label
		d.StackScriptID = script.ID
	} else if d.StackScriptID != 0 {
		script, err := client.GetStackscript(context.TODO(), d.StackScriptID)

		if err != nil {
			return fmt.Errorf("StackScript %d could not be used: %s", d.StackScriptID, err)
		}

		d.StackScriptUser = script.Username
		d.StackScriptLabel = script.Label
	}

	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Info("Creating Linode machine instance...")

	if d.SSHPort != defaultSSHPort {
		log.Infof("Using SSH port %d", d.SSHPort)
	}

	publicKey, err := d.createSSHKey()
	if err != nil {
		return err
	}

	client := d.getClient()
	boolBooted := !d.CreatePrivateIP

	// Create a linode
	createOpts := linodego.InstanceCreateOptions{
		Region:         d.Region,
		Type:           d.InstanceType,
		Label:          d.InstanceLabel,
		RootPass:       d.RootPassword,
		AuthorizedKeys: []string{strings.TrimSpace(publicKey)},
		Image:          d.InstanceImage,
		SwapSize:       &d.SwapSize,
		PrivateIP:      d.CreatePrivateIP,
		Booted:         &boolBooted,
	}

	if len(d.AuthorizedUsers) > 0 {
		createOpts.AuthorizedUsers = strings.Split(d.AuthorizedUsers, ",")
	}

	if d.Tags != "" {
		createOpts.Tags = strings.Split(d.Tags, ",")
	}

	if d.StackScriptID != 0 {
		createOpts.StackScriptID = d.StackScriptID
		createOpts.StackScriptData = d.StackScriptData
		log.Infof("Using StackScript %d: %s/%s", d.StackScriptID, d.StackScriptUser, d.StackScriptLabel)
	}

	linode, err := client.CreateInstance(context.TODO(), createOpts)
	if err != nil {
		return err
	}

	d.InstanceID = linode.ID
	d.InstanceLabel = linode.Label

	// Don't persist alias region names
	d.Region = linode.Region

	for _, address := range linode.IPv4 {
		if private := privateIP(*address); !private {
			d.IPAddress = address.String()
		} else if d.CreatePrivateIP {
			d.PrivateIPAddress = address.String()
		}
	}

	if d.IPAddress == "" {
		return errors.New("Linode IP Address is not found")
	}

	if d.CreatePrivateIP && d.PrivateIPAddress == "" {
		return errors.New("Linode Private IP Address is not found")
	}

	log.Debugf("Created Linode Instance %s (%d), IP address %q, Private IP address %q",
		d.InstanceLabel,
		d.InstanceID,
		d.IPAddress,
		d.PrivateIPAddress,
	)

	if err != nil {
		return err
	}

	if d.CreatePrivateIP {
		log.Debugf("Enabling Network Helper for Private IP configuration...")

		configs, err := client.ListInstanceConfigs(context.TODO(), linode.ID, nil)
		if err != nil {
			return err
		}
		if len(configs) == 0 {
			return fmt.Errorf("Linode Config was not found for Linode %d", linode.ID)
		}
		updateOpts := configs[0].GetUpdateOptions()
		updateOpts.Helpers.Network = true
		if _, err := client.UpdateInstanceConfig(context.TODO(), linode.ID, configs[0].ID, updateOpts); err != nil {
			return err
		}

		if err := client.BootInstance(context.TODO(), linode.ID, configs[0].ID); err != nil {
			return err
		}
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
	return ipInCIDR(ip, "10.0.0.0/8") || ipInCIDR(ip, "172.16.0.0/12") || ipInCIDR(ip, "192.168.0.0/16")
}

func ipInCIDR(ip net.IP, CIDR string) bool {
	_, ipNet, err := net.ParseCIDR(CIDR)
	if err != nil {
		log.Errorf("Error parsing CIDR %s: %s", CIDR, err)

		return false
	}
	return ipNet.Contains(ip)
}
