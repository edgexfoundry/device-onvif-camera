package driver

import (
	"fmt"
	"github.com/IOTechSystems/onvif"
	"github.com/IOTechSystems/onvif/device"
	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
	"github.com/edgexfoundry/device-onvif-camera/pkg/netscan"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	protocolName = "Onvif"
	udp          = "udp"
)

// OnvifProtocolDiscovery implements netscan.ProtocolSpecificDiscovery
type OnvifProtocolDiscovery struct {
	driver *Driver
}

func NewOnvifProtocolDiscovery(driver *Driver) *OnvifProtocolDiscovery {
	return &OnvifProtocolDiscovery{driver: driver}
}

// ProbeFilter takes in a host and a slice of ports to be scanned. It should return a slice
// of ports to actually scan, or a nil/empty slice if the host is to not be scanned at all.
// Can be used to filter out known devices from being probed again if required.
func (proto *OnvifProtocolDiscovery) ProbeFilter(_ string, ports []string) []string {
	// For onvif we do not want to do any filtering
	return ports
}

// OnConnectionDialed handles the protocol specific verification if there is actually
// a valid device or devices at the other end of the connection.
func (proto *OnvifProtocolDiscovery) OnConnectionDialed(host string, port string, conn net.Conn, params netscan.Params) ([]netscan.ProbeResult, error) {
	if devices, err := probeOnvif(host, port, params); err == nil && len(devices) > 0 {
		return mapProbeResults(host, port, devices), nil
	}

	if devices, err := probeDirect(conn, params); err == nil && len(devices) > 0 {
		return mapProbeResults(host, port, devices), nil
	}

	return nil, nil
}

// ConvertProbeResult takes a raw ProbeResult and transforms it into a
// processed DiscoveredDevice struct.
func (proto *OnvifProtocolDiscovery) ConvertProbeResult(probeResult netscan.ProbeResult, params netscan.Params) (netscan.DiscoveredDevice, error) {
	onvifDevice, ok := probeResult.Data.(onvif.Device)
	if !ok {
		return netscan.DiscoveredDevice{}, fmt.Errorf("unable to cast probe result into onvif.Device. type=%T", probeResult.Data)
	}

	discovered, err := proto.driver.createDiscoveredDevice(onvifDevice)
	if err != nil {
		return netscan.DiscoveredDevice{}, err
	}

	return netscan.DiscoveredDevice{
		Name: discovered.Name,
		Info: discovered,
	}, nil
}

func (d *Driver) createDiscoveredDevice(onvifDevice onvif.Device) (sdkModel.DiscoveredDevice, error) {
	xaddr := onvifDevice.GetDeviceParams().Xaddr
	endpointRefAddr := onvifDevice.GetDeviceParams().EndpointRefAddress
	if endpointRefAddr == "" {
		d.lc.Warnf("The EndpointRefAddress is empty from the Onvif camera, unable to add the camera %s", xaddr)
		return sdkModel.DiscoveredDevice{}, fmt.Errorf("empty EndpointRefAddress for XAddr %s", xaddr)
	}
	address, port := addressAndPort(xaddr)
	dev := contract.Device{
		// Using Xaddr as the temporary name
		Name: xaddr,
		Protocols: map[string]contract.ProtocolProperties{
			OnvifProtocol: {
				Address:    address,
				Port:       port,
				AuthMode:   d.config.DefaultAuthMode,
				SecretPath: d.config.DefaultSecretPath,
			},
		},
	}

	devInfo, edgexErr := d.getDeviceInformation(dev)
	endpointRef := endpointRefAddr
	var discovered sdkModel.DiscoveredDevice
	if edgexErr != nil {
		d.lc.Warnf("failed to get the device information for the camera %s, %v", endpointRef, edgexErr)
		dev.Protocols[OnvifProtocol][SecretPath] = endpointRef
		discovered = sdkModel.DiscoveredDevice{
			Name:        endpointRef,
			Protocols:   dev.Protocols,
			Description: "Auto discovered Onvif camera",
			Labels:      []string{"auto-discovery"},
		}
		d.lc.Debugf("Discovered unknown camera from the address '%s'", xaddr)
	} else {
		dev.Protocols[OnvifProtocol][Manufacturer] = devInfo.Manufacturer
		dev.Protocols[OnvifProtocol][Model] = devInfo.Model
		dev.Protocols[OnvifProtocol][FirmwareVersion] = devInfo.FirmwareVersion
		dev.Protocols[OnvifProtocol][SerialNumber] = devInfo.SerialNumber
		dev.Protocols[OnvifProtocol][HardwareId] = devInfo.HardwareId

		// Spaces are not allowed in the device name
		deviceName := fmt.Sprintf("%s-%s-%s",
			strings.ReplaceAll(devInfo.Manufacturer, " ", "-"),
			strings.ReplaceAll(devInfo.Model, " ", "-"),
			endpointRefAddr)

		discovered = sdkModel.DiscoveredDevice{
			Name:        deviceName,
			Protocols:   dev.Protocols,
			Description: fmt.Sprintf("%s %s Camera", devInfo.Manufacturer, devInfo.Model),
			Labels:      []string{"auto-discovery", devInfo.Manufacturer, devInfo.Model},
		}
		d.lc.Debugf("Discovered camera from the address '%s'", xaddr)
	}
	return discovered, nil
}

func mapProbeResults(host, port string, devices []onvif.Device) (res []netscan.ProbeResult) {
	for _, dev := range devices {
		res = append(res, netscan.ProbeResult{
			Host: host,
			Port: port,
			Data: dev,
		})
	}
	return res
}

func parseProbeResponse(conn net.Conn, params netscan.Params) ([]onvif.Device, error) {
	addr := conn.RemoteAddr().String()
	if err := conn.SetReadDeadline(time.Now().Add(params.Timeout)); err != nil {
		err = errors.Wrapf(err, "%s: failed to set read deadline", addr)
		params.Logger.Debugf(err.Error())
		return nil, err
	}

	buf := make([]byte, 5)
	if _, err := io.ReadFull(conn, buf); err != nil {
		if params.NetworkProtocol == udp {
			// on udp connections all timeouts result in this
			return nil, nil
		}
		err = errors.Wrapf(err, "%s: failed to read header", addr)
		params.Logger.Debugf(err.Error())
		return nil, err
	}
	if string(buf) != "<?xml" {
		params.Logger.Debugf("%s: non xml response received", addr)
		return nil, nil
	}

	params.Logger.Infof("%s: got xml", addr)

	buf2, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	response := string(buf) + string(buf2)
	params.Logger.Infof("%s: Got Bytes: %s", addr, response)

	devices, err := wsdiscovery.DevicesFromProbeResponses([]string{response})
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		params.Logger.Infof("%s: no devices matched from probe response", addr)
		return nil, nil
	}

	return devices, nil
}

func probeDirect(conn net.Conn, params netscan.Params) ([]onvif.Device, error) {
	probeSOAP := wsdiscovery.BuildProbeMessage(uuid.Must(uuid.NewV4()).String(), nil, nil,
		map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})
	if _, err := conn.Write([]byte(probeSOAP.String())); err != nil {
		err = errors.Wrap(err, "failed to write probe message")
		params.Logger.Debugf(err.Error())
		return nil, err
	}

	return parseProbeResponse(conn, params)
}

func probeOnvif(host, port string, params netscan.Params) ([]onvif.Device, error) {
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr: fmt.Sprintf("%s:%s", host, port),
		HttpClient: &http.Client{
			Timeout: params.Timeout,
		},
	})
	if err != nil {
		err = errors.Wrap(err, "failed to create potential onvif device")
		params.Logger.Debugf(err.Error())
		return nil, err
	}

	res, err := dev.CallOnvifFunction(onvif.DeviceWebService, onvif.GetEndpointReference, nil)
	if err != nil {
		err = errors.Wrap(err, "failed to call GetEndpointReference")
		params.Logger.Debugf(err.Error())
		return nil, err
	}

	ref := res.(*device.GetEndpointReferenceResponse)
	dp := dev.GetDeviceParams()
	uuidElements := strings.Split(ref.GUID, ":")
	dp.EndpointRefAddress = uuidElements[len(uuidElements)-1]
	nvt, err := onvif.NewDevice(dp)
	if err != nil {
		err = errors.Wrap(err, "failed to create new onvif device from old device")
		params.Logger.Debugf(err.Error())
		return nil, err
	}
	return []onvif.Device{*nvt}, nil
}

// makeDeviceMap creates a lookup table of existing devices by tcp address in order to skip scanning
func (d *Driver) makeDeviceMap() map[string]contract.Device {
	devices := d.svc.Devices()
	deviceMap := make(map[string]contract.Device, len(devices))

	for _, dev := range devices {
		if dev.Name == d.serviceName {
			// skip control plane device
			continue
		}

		onvifInfo := dev.Protocols[protocolName]
		if onvifInfo == nil {
			d.lc.Warnf("Found registered device %s without %s protocol information.", dev.Name, protocolName)
			continue
		}

		host, port := onvifInfo["Address"], onvifInfo["Port"]
		if host == "" || port == "" {
			d.lc.Warnf("Registered device is missing required %s protocol information. Address: %v, Port: %v",
				protocolName, host, port)
			continue
		}

		deviceMap[host+":"+port] = dev
	}

	return deviceMap
}
