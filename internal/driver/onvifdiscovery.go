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
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	protocolName = "Onvif"
	bufSize      = 8192
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

func probeOnvif(host string, port string, params netscan.Params) ([]netscan.ProbeResult, error) {
	// attempt to create an Onvif connection with the device
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr: fmt.Sprintf("%s:%s", host, port),
		HttpClient: &http.Client{
			Timeout: params.Timeout,
		},
	})
	// if this failed, no need to try anything else, just bail out now
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create potential onvif device for %s:%s", host, port)
	}

	// attempt to determine the EndpointReferenceAddress UUID
	res, err := dev.CallOnvifFunction(onvif.DeviceWebService, onvif.GetEndpointReference, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call GetEndpointReference for %s:%s", host, port)
	}

	ref, ok := res.(*device.GetEndpointReferenceResponse)
	if !ok {
		return nil, fmt.Errorf("unable to cast response to GetEndpointReferenceResponse for %s:%s. type=%T", host, port, res)
	}
	dp := dev.GetDeviceParams()
	uuidElements := strings.Split(ref.GUID, ":")
	dp.EndpointRefAddress = uuidElements[len(uuidElements)-1]
	// todo: there has got to be a smarter way to do this than creating a new device
	nvt, err := onvif.NewDevice(dp)
	if err != nil {
		// this shouldn't ever happen
		err = errors.Wrap(err, "failed to create new onvif device from old device")
		params.Logger.Errorf(err.Error())
		return nil, err
	}
	return mapProbeResults(host, port, []onvif.Device{*nvt}), nil
}

func probeRaw(host string, port string, conn net.Conn, params netscan.Params) ([]netscan.ProbeResult, error) {
	// attempt a basic direct probe approach using the open connection
	devices, err := executeRawProbe(conn, params)
	if err != nil {
		params.Logger.Debugf(err.Error())
	} else if len(devices) > 0 {
		return mapProbeResults(host, port, devices), nil
	}
	return nil, err
}

// OnConnectionDialed handles the protocol specific verification if there is actually
// a valid device or devices at the other end of the connection.
func (proto *OnvifProtocolDiscovery) OnConnectionDialed(host string, port string, conn net.Conn, params netscan.Params) ([]netscan.ProbeResult, error) {
	//res, err := probeOnvif(host, port, params)
	//if err != nil {
	//	params.Logger.Errorf(err.Error())
	//} else if len(res) > 0 {
	//	return res, nil
	//}

	return probeRaw(host, port, conn, params)
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
	dev.Protocols[OnvifProtocol][EndpointRefAddress] = endpointRef
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

func executeRawProbe(conn net.Conn, params netscan.Params) ([]onvif.Device, error) {
	probeSOAP := wsdiscovery.BuildProbeMessage(uuid.Must(uuid.NewV4()).String(), nil, nil,
		map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})

	addr := conn.RemoteAddr().String()
	if err := conn.SetDeadline(time.Now().Add(params.Timeout)); err != nil {
		return nil, errors.Wrapf(err, "%s: failed to set read/write deadline", addr)
	}

	if _, err := conn.Write([]byte(probeSOAP.String())); err != nil {
		return nil, errors.Wrap(err, "failed to write probe message")
	}

	var responses []string
	buf := make([]byte, bufSize)

	for {
		n, _, err := (conn.(net.PacketConn)).ReadFrom(buf)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				params.Logger.Warnf(err.Error())
			}
			break
		}
		responses = append(responses, string(buf[0:n]))
	}

	//buf2, err := io.ReadAll(conn)
	//if err != nil {
	//	params.Logger.Debugf("%s: Got Bytes: %s", addr, string(buf2))
	//	return nil, err
	//}
	//response := strings.Join(result, "")
	//response := string(buf) + string(buf2)

	if len(responses) == 0 {
		params.Logger.Tracef("%s: No Response", addr)
		return nil, nil
	}
	for i, resp := range responses {
		params.Logger.Debugf("%s: Response %d of %d: %s", addr, i+1, len(responses), resp)
	}

	devices, err := wsdiscovery.DevicesFromProbeResponses(responses)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		params.Logger.Debugf("%s: no devices matched from probe response", addr)
		return nil, nil
	}

	return devices, nil
}

// makeDeviceMap creates a lookup table of existing devices by EndpointRefAddress
// todo: will be used in the future for device re-discovery purposes
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

		endpointRef := onvifInfo["EndpointRefAddress"]
		if endpointRef == "" {
			d.lc.Warnf("Registered device %s is missing required %s protocol information: EndpointRefAddress.",
				dev.Name, protocolName)
			continue
		}

		deviceMap[endpointRef] = dev
	}

	return deviceMap
}
