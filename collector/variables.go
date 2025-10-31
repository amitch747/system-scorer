package collector

import (
	"regexp"
)

// '^(veth|cni|flannel|docker|br-).*'
var NetDeviceFilter = regexp.MustCompile("^(lo|veth|docker|br-|tun).*")

var ScrapeInterval float64 = 15
