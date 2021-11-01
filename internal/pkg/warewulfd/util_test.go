package warewulfd

import (
	"testing"
)

func Test_CidrRangeContains(t *testing.T) {
	runCidrRangeContains(t, "128.128.128.0/24", "128.128.128.128", true)
	runCidrRangeContains(t, "192.0.2.0/24", "::ffff:192.0.2.128", true)
	runCidrRangeContains(t, "10.0.0.0/8", "10.1.1.1", true)
	runCidrRangeContains(t, "10.0.0.0/8", "11.1.1.1", false)
	runCidrRangeContains(t, "10.0.0.0/16", "10.0.1.1", true)
	runCidrRangeContains(t, "10.0.0.0/16", "10.1.1.1", false)
	runCidrRangeContains(t, "10.0.0.0/24", "10.0.0.1", true)
	runCidrRangeContains(t, "10.0.0.0/24", "10.0.1.1", false)
	runCidrRangeContains(t, "2001:0db8:85a3:0000:0000:8a2e:0370:7334/24", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true)
	runCidrRangeContains(t, "192.0.2.0/24", "foobar", false)
	runCidrRangeContains(t, "invalid", "10.0.0.1", false)
}

func runCidrRangeContains(t *testing.T, cidrRange string, checkIP string, assert bool) {
	res, _ := cidrRangeContains(cidrRange, checkIP)
	if res != assert {
		t.Errorf("Assertion (have: %v should be: %v) failed on cidrRange %s with checkIP %s", res, assert, cidrRange, checkIP)
	}
}
