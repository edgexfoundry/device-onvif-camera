package main

import (
	"fmt"
	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"time"
)

const (
	bufSize = 8192
)

func main() {
	//if len(os.Args) < 2 {
	//	fmt.Printf("usage: %s <interface-name>\n", os.Args[0])
	//	os.Exit(2)
	//}
	//wsdiscovery.GetAvailableDevicesAtSpecificEthernetInterface(os.Args[1])

	probeSOAP := wsdiscovery.BuildProbeMessage(uuid.Must(uuid.NewV4()).String(), nil, nil,
		map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})

	res := SendUDPUnicast(probeSOAP.String(), net.IPv4(10, 0, 0, 217))
	log.Printf("%v", res)

	//res := SendUDPMulticast(probeSOAP.String())
	//log.Printf("%v", res)
}

func SendUDPUnicast(msg string, ip net.IP) []string {
	var result []string
	data := []byte(msg)

	c, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(nil, &net.UDPAddr{IP: ip}); err != nil {
		fmt.Println(err)
	}

	dst := &net.UDPAddr{IP: ip, Port: 3702}
	if _, err := p.WriteTo(data, nil, dst); err != nil {
		fmt.Println(err)
	}

	if err := p.SetReadDeadline(time.Now().Add(time.Second * 1)); err != nil {
		log.Fatal(err)
	}

	for {
		b := make([]byte, bufSize)
		n, _, _, err := p.ReadFrom(b)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println(err)
			}
			break
		}
		result = append(result, string(b[0:n]))
	}
	return result
}

func SendUDPMulticast(msg string) []string {
	var result []string
	data := []byte(msg)

	group := net.IPv4(239, 255, 255, 250)

	c, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(nil, &net.UDPAddr{IP: group}); err != nil {
		fmt.Println(err)
	}

	dst := &net.UDPAddr{IP: group, Port: 3702}
	p.SetMulticastTTL(2)
	if _, err := p.WriteTo(data, nil, dst); err != nil {
		fmt.Println(err)
	}

	if err := p.SetReadDeadline(time.Now().Add(time.Second * 1)); err != nil {
		log.Fatal(err)
	}

	for {
		b := make([]byte, bufSize)
		n, _, _, err := p.ReadFrom(b)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println(err)
			}
			break
		}
		result = append(result, string(b[0:n]))
	}
	return result
}
