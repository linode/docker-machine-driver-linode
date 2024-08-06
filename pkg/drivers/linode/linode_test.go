package linode

import (
	"net"
	"reflect"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/google/go-cmp/cmp"
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
