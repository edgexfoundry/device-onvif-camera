package onvif

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/IOTechSystems/onvif/device"
	"github.com/IOTechSystems/onvif/gosoap"
	"github.com/IOTechSystems/onvif/xsd/onvif"

	"github.com/beevik/etree"
)

// Xlmns XML Scheam
var Xlmns = map[string]string{
	"onvif":   "http://www.onvif.org/ver10/schema",
	"tds":     "http://www.onvif.org/ver10/device/wsdl",
	"trt":     "http://www.onvif.org/ver10/media/wsdl",
	"tr2":     "http://www.onvif.org/ver20/media/wsdl",
	"tev":     "http://www.onvif.org/ver10/events/wsdl",
	"tptz":    "http://www.onvif.org/ver20/ptz/wsdl",
	"timg":    "http://www.onvif.org/ver20/imaging/wsdl",
	"tan":     "http://www.onvif.org/ver20/analytics/wsdl",
	"xmime":   "http://www.w3.org/2005/05/xmlmime",
	"wsnt":    "http://docs.oasis-open.org/wsn/b-2",
	"xop":     "http://www.w3.org/2004/08/xop/include",
	"wsa":     "http://www.w3.org/2005/08/addressing",
	"wstop":   "http://docs.oasis-open.org/wsn/t-1",
	"wsntw":   "http://docs.oasis-open.org/wsn/bw-2",
	"wsrf-rw": "http://docs.oasis-open.org/wsrf/rw-2",
	"wsaw":    "http://www.w3.org/2006/05/addressing/wsdl",
	"tt":      "http://www.onvif.org/ver10/recording/wsdl",
	"wsse":    "http://docs.oasis-open.org/wss/2004/01/oasis-200401",
	"wsu":     "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd",
}

// DeviceType alias for int
type DeviceType int

// Onvif Device Tyoe
const (
	NVD DeviceType = iota
	NVS
	NVA
	NVT

	ContentType = "Content-Type"
)

func (devType DeviceType) String() string {
	stringRepresentation := []string{
		"NetworkVideoDisplay",
		"NetworkVideoStorage",
		"NetworkVideoAnalytics",
		"NetworkVideoTransmitter",
	}
	i := uint8(devType)
	switch {
	case i <= uint8(NVT):
		return stringRepresentation[i]
	default:
		return strconv.Itoa(int(i))
	}
}

// DeviceInfo struct contains general information about ONVIF device
type DeviceInfo struct {
	Name            string
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareId      string
}

// Device for a new device of onvif and DeviceInfo
// struct represents an abstract ONVIF device.
// It contains methods, which helps to communicate with ONVIF device
type Device struct {
	params       DeviceParams
	endpoints    map[string]string
	info         DeviceInfo
	digestClient *DigestClient
}

type DeviceParams struct {
	Xaddr              string
	EndpointRefAddress string
	Username           string
	Password           string
	HttpClient         *http.Client
	AuthMode           string
}

// GetServices return available endpoints
func (dev *Device) GetServices() map[string]string {
	return dev.endpoints
}

// GetServices return available endpoints
func (dev *Device) GetDeviceInfo() DeviceInfo {
	return dev.info
}

// SetDeviceInfoFromScopes goes through the scopes and sets the device info fields for supported categories (currently name and hardware).
// See 7.3.2.2 Scopes in the ONVIF Core Specification (https://www.onvif.org/specs/core/ONVIF-Core-Specification.pdf).
func (dev *Device) SetDeviceInfoFromScopes(scopes []string) {
	newInfo := dev.info
	supportedScopes := []struct {
		category string
		setField func(s string)
	}{
		{category: "name", setField: func(s string) { newInfo.Name = s }},
		{category: "hardware", setField: func(s string) { newInfo.Model = s }},
	}

	for _, s := range scopes {
		for _, supp := range supportedScopes {
			fullScope := fmt.Sprintf("onvif://www.onvif.org/%s/", supp.category)
			scopeValue, matchesScope := strings.CutPrefix(s, fullScope)
			if matchesScope {
				unescaped, err := url.QueryUnescape(scopeValue)
				if err != nil {
					continue
				}
				supp.setField(unescaped)
			}
		}
	}
	dev.info = newInfo
}

func (dev *Device) getSupportedServices(resp *http.Response) {
	doc := etree.NewDocument()

	data, _ := io.ReadAll(resp.Body)

	if err := doc.ReadFromBytes(data); err != nil {
		//log.Println(err.Error())
		return
	}
	services := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/*/XAddr")
	for _, j := range services {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}

	extensionServices := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/Extension/*/XAddr")
	for _, j := range extensionServices {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}
}

// NewDevice function construct a ONVIF Device entity
func NewDevice(params DeviceParams) (*Device, error) {
	dev := new(Device)
	dev.params = params
	dev.endpoints = make(map[string]string)
	dev.addEndpoint("Device", "http://"+dev.params.Xaddr+"/onvif/device_service")

	if dev.params.HttpClient == nil {
		dev.params.HttpClient = new(http.Client)
	}
	dev.digestClient = NewDigestClient(dev.params.HttpClient, dev.params.Username, dev.params.Password)

	getCapabilities := device.GetCapabilities{Category: []onvif.CapabilityCategory{"All"}}

	resp, err := dev.CallMethod(getCapabilities)

	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("camera is not available at " + dev.params.Xaddr + " or it does not support ONVIF services")
	}

	dev.getSupportedServices(resp)
	return dev, nil
}

func (dev *Device) addEndpoint(Key, Value string) {
	//use lowCaseKey
	//make key having ability to handle Mixed Case for Different vendor devcie (e.g. Events EVENTS, events)
	lowCaseKey := strings.ToLower(Key)

	// Replace host with host from device params.
	if u, err := url.Parse(Value); err == nil {
		u.Host = dev.params.Xaddr
		Value = u.String()
	}

	dev.endpoints[lowCaseKey] = Value

	if lowCaseKey == strings.ToLower(MediaWebService) {
		// Media2 uses the same endpoint but different XML name space
		dev.endpoints[strings.ToLower(Media2WebService)] = Value
	}
}

// GetEndpoint returns specific ONVIF service endpoint address
func (dev *Device) GetEndpoint(name string) string {
	return dev.endpoints[name]
}

// getEndpoint functions get the target service endpoint in a better way
func (dev *Device) getEndpoint(endpoint string) (string, error) {

	// common condition, endpointMark in map we use this.
	if endpointURL, bFound := dev.endpoints[endpoint]; bFound {
		return endpointURL, nil
	}

	//but ,if we have endpoint like event、analytic
	//and sametime the Targetkey like : events、analytics
	//we use fuzzy way to find the best match url
	var endpointURL string
	for targetKey := range dev.endpoints {
		if strings.Contains(targetKey, endpoint) {
			endpointURL = dev.endpoints[targetKey]
			return endpointURL, nil
		}
	}
	return endpointURL, errors.New("target endpoint service not found")
}

// CallMethod functions call an method, defined <method> struct.
// You should use Authenticate method to call authorized requests.
func (dev *Device) CallMethod(method interface{}) (*http.Response, error) {
	pkgPath := strings.Split(reflect.TypeOf(method).PkgPath(), "/")
	pkg := strings.ToLower(pkgPath[len(pkgPath)-1])

	endpoint, err := dev.getEndpoint(pkg)
	if err != nil {
		return nil, err
	}
	requestBody, err := xml.Marshal(method)
	if err != nil {
		return nil, err
	}
	return dev.SendSoap(endpoint, string(requestBody))
}

func (dev *Device) GetDeviceParams() DeviceParams {
	return dev.params
}

func (dev *Device) GetEndpointByRequestStruct(requestStruct interface{}) (string, error) {
	pkgPath := strings.Split(reflect.TypeOf(requestStruct).Elem().PkgPath(), "/")
	pkg := strings.ToLower(pkgPath[len(pkgPath)-1])

	endpoint, err := dev.getEndpoint(pkg)
	if err != nil {
		return "", err
	}
	return endpoint, err
}

func (dev *Device) SendSoap(endpoint string, xmlRequestBody string) (resp *http.Response, err error) {
	soap := gosoap.NewEmptySOAP()
	soap.AddStringBodyContent(xmlRequestBody)
	soap.AddRootNamespaces(Xlmns)
	if dev.params.AuthMode == UsernameTokenAuth || dev.params.AuthMode == Both {
		err = soap.AddWSSecurity(dev.params.Username, dev.params.Password)
		if err != nil {
			return nil, fmt.Errorf("send soap request failed: %w", err)
		}
	}

	if dev.params.AuthMode == DigestAuth || dev.params.AuthMode == Both {
		resp, err = dev.digestClient.Do(http.MethodPost, endpoint, soap.String())
	} else {
		var req *http.Request
		req, err = createHttpRequest(http.MethodPost, endpoint, soap.String())
		if err != nil {
			return nil, err
		}
		resp, err = dev.params.HttpClient.Do(req)
	}
	return resp, err
}

func createHttpRequest(httpMethod string, endpoint string, soap string) (req *http.Request, err error) {
	req, err = http.NewRequest(httpMethod, endpoint, bytes.NewBufferString(soap))
	if err != nil {
		return nil, err
	}
	req.Header.Set(ContentType, "application/soap+xml; charset=utf-8")
	return req, nil
}

func (dev *Device) CallOnvifFunction(serviceName, functionName string, data []byte) (interface{}, error) {
	function, err := FunctionByServiceAndFunctionName(serviceName, functionName)
	if err != nil {
		return nil, err
	}
	request, err := createRequest(function, data)
	if err != nil {
		return nil, fmt.Errorf("fail to create '%s' request for the web service '%s', %v", functionName, serviceName, err)
	}

	endpoint, err := dev.GetEndpointByRequestStruct(request)
	if err != nil {
		return nil, err
	}

	requestBody, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}
	xmlRequestBody := string(requestBody)

	servResp, err := dev.SendSoap(endpoint, xmlRequestBody)
	if err != nil {
		return nil, fmt.Errorf("fail to send the '%s' request for the web service '%s', %v", functionName, serviceName, err)
	}
	defer servResp.Body.Close()

	rsp, err := io.ReadAll(servResp.Body)
	if err != nil {
		return nil, err
	}

	responseEnvelope, err := createResponse(function, rsp)
	if err != nil {
		return nil, fmt.Errorf("fail to create '%s' response for the web service '%s', %v", functionName, serviceName, err)
	}

	if servResp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("fail to verify the authentication for the function '%s' of web service '%s'. Onvif error: %s",
			functionName, serviceName, responseEnvelope.Body.Fault.String())
	} else if servResp.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("invalid request for the function '%s' of web service '%s'. Onvif error: %s",
			functionName, serviceName, responseEnvelope.Body.Fault.String())
	} else if servResp.StatusCode > http.StatusNoContent {
		return nil, fmt.Errorf("fail to execute the request for the function '%s' of web service '%s'. Onvif error: %s",
			functionName, serviceName, responseEnvelope.Body.Fault.String())
	}
	return responseEnvelope.Body.Content, nil
}

func createRequest(function Function, data []byte) (interface{}, error) {
	request := function.Request()
	if len(data) > 0 {
		err := json.Unmarshal(data, request)
		if err != nil {
			return nil, err
		}
	}
	return request, nil
}

func createResponse(function Function, data []byte) (*gosoap.SOAPEnvelope, error) {
	response := function.Response()
	responseEnvelope := gosoap.NewSOAPEnvelope(response)
	err := xml.Unmarshal(data, responseEnvelope)
	if err != nil {
		return nil, err
	}
	return responseEnvelope, nil
}

// SendGetSnapshotRequest sends the Get request to retrieve the snapshot from the Onvif camera
// The parameter url is come from the "GetSnapshotURI" command.
func (dev *Device) SendGetSnapshotRequest(url string) (resp *http.Response, err error) {
	soap := gosoap.NewEmptySOAP()
	soap.AddRootNamespaces(Xlmns)
	if dev.params.AuthMode == UsernameTokenAuth {
		err = soap.AddWSSecurity(dev.params.Username, dev.params.Password)
		if err != nil {
			return nil, fmt.Errorf("send GetSnapshotRequest failed: %w", err)
		}
		var req *http.Request
		req, err = createHttpRequest(http.MethodGet, url, soap.String())
		if err != nil {
			return nil, err
		}
		// Basic auth might work for some camera
		req.SetBasicAuth(dev.params.Username, dev.params.Password)
		resp, err = dev.params.HttpClient.Do(req)

	} else if dev.params.AuthMode == DigestAuth || dev.params.AuthMode == Both {
		err = soap.AddWSSecurity(dev.params.Username, dev.params.Password)
		if err != nil {
			return nil, fmt.Errorf("send GetSnapshotRequest failed: %w", err)
		}
		resp, err = dev.digestClient.Do(http.MethodGet, url, soap.String())

	} else {
		var req *http.Request
		req, err = createHttpRequest(http.MethodGet, url, soap.String())
		if err != nil {
			return nil, err
		}
		resp, err = dev.params.HttpClient.Do(req)
	}
	return resp, err
}
