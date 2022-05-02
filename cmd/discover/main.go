package main

import (
	"fmt"
	"github.com/IOTechSystems/onvif/ws-discovery"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <interface-name>\n", os.Args[0])
		os.Exit(2)
	}
	wsdiscovery.GetAvailableDevicesAtSpecificEthernetInterface(os.Args[1])
}
