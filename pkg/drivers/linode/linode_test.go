package linode

import (
	"net"
	"reflect"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/google/go-cmp/cmp"
	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":     "PROJECT",
			"linode-root-pass": "ROOTPASS",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetConfigFromFlagsInterfaceRequiresVPC(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":          "PROJECT",
			"linode-use-interfaces": true,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires --linode-vpc-subnet-id")
}

func TestSetConfigFromFlagsInterfaceConflictsWithLegacyPrivateIP(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":             "PROJECT",
			"linode-use-interfaces":    true,
			"linode-vpc-subnet-id":     456,
			"linode-create-private-ip": true,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "linode-use-interfaces")
}

func TestSetConfigFromFlagsInterfaceHappyPath(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":                        "PROJECT",
			"linode-use-interfaces":               true,
			"linode-vpc-subnet-id":                456,
			"linode-vpc-private-ip":               "10.0.0.10",
			"linode-vpc-interface-firewall-id":    321,
			"linode-public-interface-firewall-id": 789,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)
	assert.True(t, driver.UseInterfaces)
	assert.Equal(t, 456, driver.VPCSubnetID)
	assert.Equal(t, "10.0.0.10", driver.VPCPrivateIP)
	assert.Equal(t, 321, driver.VPCInterfaceFirewallID)
	assert.Equal(t, 789, driver.PublicInterfaceFirewallID)
}

func TestSetConfigFromFlagsInterfaceFirewallRequiresInterfaces(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":                        "PROJECT",
			"linode-public-interface-firewall-id": 111,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "linode-use-interfaces")
}

func TestSetConfigFromFlagsInterfaceFirewallMustBeNonNegative(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":                        "PROJECT",
			"linode-use-interfaces":               true,
			"linode-vpc-subnet-id":                456,
			"linode-public-interface-firewall-id": -2,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--linode-public-interface-firewall-id")
}

func TestSetConfigFromFlagsVPCInterfaceFirewallRequiresInterfaces(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":                     "PROJECT",
			"linode-vpc-interface-firewall-id": 222,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "linode-use-interfaces")
}

func TestSetConfigFromFlagsVPCInterfaceFirewallMustBeNonNegative(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"linode-token":                     "PROJECT",
			"linode-use-interfaces":            true,
			"linode-vpc-subnet-id":             456,
			"linode-vpc-interface-firewall-id": -2,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--linode-vpc-interface-firewall-id")
}

func TestPrivateIP(t *testing.T) {
	ip := net.IP{}
	for _, addr := range [][]byte{
		[]byte("172.16.0.1"),
		[]byte("192.168.0.1"),
		[]byte("10.0.0.1"),
	} {
		if err := ip.UnmarshalText(addr); err != nil {
			t.Error(err)
		}
		assert.True(t, privateIP(ip))
	}

	if err := ip.UnmarshalText([]byte("1.1.1.1")); err != nil {
		t.Error(err)
	}
	assert.False(t, privateIP(ip))
}

func TestIPInCIDR(t *testing.T) {
	tenOne := net.IP{}

	if err := tenOne.UnmarshalText([]byte("10.0.0.1")); err != nil {
		t.Error(err)
	}
	assert.True(t, ipInCIDR(tenOne, "10.0.0.0/8"), "10.0.0.1 is in 10.0.0.0/8")
	assert.False(t, ipInCIDR(tenOne, "254.0.0.0/8"), "10.0.0.1 is not in 254.0.0.0/8")
}

func TestNormalizeInstanceLabel(t *testing.T) {
	inputLabel := "_mycoollabel25';./__----=][[this,label,is,really[good]and]long[wow+that'scrazy[]what[a\\good!labelname."
	expectedResult := "mycoollabel25._-thislabelisreallygoodandlongwowthatscrazywhatago"

	result, err := normalizeInstanceLabel(inputLabel)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatal(cmp.Diff(result, expectedResult))
	}
}

func TestFirstVPCIPv4SkipsRanges(t *testing.T) {
	ip := "10.0.0.5"
	ipRange := "10.0.0.0/24"

	got := firstVPCIPv4([]*linodego.VPCIP{
		{AddressRange: &ipRange},
		{Address: &ip},
	})

	assert.Equal(t, ip, got)
}
