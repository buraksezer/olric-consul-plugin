package main

import (
	"github.com/buraksezer/olric-consul-plugin/lib"
)

// ServiceDiscovery defines a service discovery plugin
//for Olric, backed by Consul.
var ServiceDiscovery lib.ConsulDiscovery

func main() {
	_ = ServiceDiscovery
}
