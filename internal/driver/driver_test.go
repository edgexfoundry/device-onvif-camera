// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/IOTechSystems/onvif"
	"github.com/IOTechSystems/onvif/xsd"
	xsdOnvif "github.com/IOTechSystems/onvif/xsd/onvif"
	"github.com/stretchr/testify/mock"

	"github.com/IOTechSystems/onvif/device"
	sdkMocks "github.com/edgexfoundry/device-sdk-go/v2/pkg/interfaces/mocks"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDeviceName = "test-device"
	getFunction    = "getFunction"
)

var (
	ptrTrue  = boolPointer(true)
	ptrFalse = boolPointer(false)
)

func boolPointer(val bool) *xsd.Boolean {
	b := xsd.Boolean(val)
	return &b
}

func createDriverWithMockService() (*Driver, *sdkMocks.DeviceServiceSDK) {
	mockService := &sdkMocks.DeviceServiceSDK{}
	driver := &Driver{sdkService: mockService, lc: logger.MockLogger{}}
	return driver, mockService
}

func createTestDevice() models.Device {
	return models.Device{Name: testDeviceName, Protocols: map[string]models.ProtocolProperties{
		OnvifProtocol: map[string]string{
			DeviceStatus: Unreachable,
		},
	}}
}

func createTestDeviceWithProtocols(protocols map[string]models.ProtocolProperties) models.Device {
	return models.Device{Name: testDeviceName, Protocols: protocols}
}

func TestParametersFromURLRawQuery(t *testing.T) {
	parameters := `{ "ProfileToken": "Profile_1" }`
	base64EncodedStr := base64.StdEncoding.EncodeToString([]byte(parameters))
	req := sdkModel.CommandRequest{
		Attributes: map[string]interface{}{
			URLRawQuery: fmt.Sprintf("%s=%s", jsonObject, base64EncodedStr),
		},
	}
	data, err := parametersFromURLRawQuery(req)
	require.NoError(t, err)
	assert.Equal(t, parameters, string(data))
}

// TestAddressAndPort verifies splitting of address and port from a given string.
func TestAddressAndPort(t *testing.T) {

	tests := []struct {
		input           string
		expectedAddress string
		expectedPort    string
	}{
		{
			input:           "localhost:80",
			expectedAddress: "localhost",
			expectedPort:    "80",
		},
		{
			input:           "localhost",
			expectedAddress: "localhost",
			expectedPort:    "80",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			resultAddress, resultPort := addressAndPort(test.input)
			assert.Equal(t, test.expectedAddress, resultAddress)
			assert.Equal(t, test.expectedPort, resultPort)
		})
	}
}

func TestDriver_HandleReadCommands(t *testing.T) {
	driver, mockService := createDriverWithMockService()
	driver.clientsMu = new(sync.RWMutex)

	tests := []struct {
		name          string
		deviceName    string
		protocols     map[string]models.ProtocolProperties
		reqs          []sdkModel.CommandRequest
		resp          string
		data          string
		expected      []*sdkModel.CommandValue
		errorExpected bool
	}{
		{
			name:       "simple read for RebootNeeded",
			deviceName: testDeviceName,
			reqs: []sdkModel.CommandRequest{
				{
					DeviceResourceName: RebootNeeded,
					Attributes: map[string]interface{}{
						getFunction: RebootNeeded,
						"service":   EdgeXWebService,
					},
					Type: "Bool",
				}},
			expected: []*sdkModel.CommandValue{
				{
					DeviceResourceName: RebootNeeded,
					Type:               "Bool",
					Value:              false,
					Tags:               map[string]string{},
				}},
		},
		{
			name:       "simple read of DeviceInformation",
			deviceName: testDeviceName,
			reqs: []sdkModel.CommandRequest{
				{
					DeviceResourceName: "DeviceInformation",
					Attributes: map[string]interface{}{
						getFunction: "GetDeviceInformation",
						"service":   onvif.DeviceWebService,
					},
					Type: "Object",
				}},
			resp: `<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Header />
  <Body>
    <Content>
      <Manufacturer>Intel</Manufacturer>
      <Model>SimCamera</Model>
      <FirmwareVersion>2.4a</FirmwareVersion>
      <SerialNumber>46d1ab8d</SerialNumber>
      <HardwareId>1.0</HardwareId>
    </Content>
  </Body>
</Envelope>`,
			expected: []*sdkModel.CommandValue{
				{
					DeviceResourceName: "DeviceInformation",
					Type:               "Object",
					Value: &device.GetDeviceInformationResponse{
						Manufacturer:    "Intel",
						Model:           "SimCamera",
						FirmwareVersion: "2.4a",
						SerialNumber:    "46d1ab8d",
						HardwareId:      "1.0",
					},
					Tags: map[string]string{},
				}},
		},
		{
			name:       "simple read of GetNetworkInterfaces",
			deviceName: testDeviceName,
			reqs: []sdkModel.CommandRequest{
				{
					DeviceResourceName: "NetworkInterfaces",
					Attributes: map[string]interface{}{
						getFunction: "GetNetworkInterfaces",
						"service":   onvif.DeviceWebService,
					},
					Type: "Object",
				}},
			resp: `<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Header />
  <Body>
    <Content>
      <NetworkInterfaces token="NET_TOKEN_4047201479">
        <Enabled>true</Enabled>
        <Info>
          <Name>eth0</Name>
          <HwAddress>02:42:C0:A8:90:0E</HwAddress>
          <MTU>1500</MTU>
        </Info>
        <IPv4>
          <Enabled>true</Enabled>
          <Config>
            <Manual>
              <Address>192.168.144.14</Address>
              <PrefixLength>20</PrefixLength>
            </Manual>
            <DHCP>false</DHCP>
          </Config>
        </IPv4>
      </NetworkInterfaces>
    </Content>
  </Body>
</Envelope>`,
			expected: []*sdkModel.CommandValue{
				{
					DeviceResourceName: "NetworkInterfaces",
					Type:               "Object",
					Value: &device.GetNetworkInterfacesResponse{
						NetworkInterfaces: xsdOnvif.NetworkInterface{
							DeviceEntity: xsdOnvif.DeviceEntity{
								Token: "NET_TOKEN_4047201479",
							},
							Enabled: ptrTrue,
							Info: &xsdOnvif.NetworkInterfaceInfo{
								Name:      "eth0",
								HwAddress: "02:42:C0:A8:90:0E",
								MTU:       1500,
							},
							IPv4: &xsdOnvif.IPv4NetworkInterface{
								Enabled: ptrTrue,
								Config: &xsdOnvif.IPv4Configuration{
									Manual: &xsdOnvif.PrefixedIPv4Address{
										Address:      "192.168.144.14",
										PrefixLength: 20,
									},
									DHCP: ptrFalse,
								},
							},
						},
					},
					Tags: map[string]string{},
				}},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				_, err := writer.Write([]byte(test.resp))
				assert.NoError(t, err)
			}))
			defer server.Close()

			client, mockDevice := createOnvifClientWithMockDevice(driver, testDeviceName)
			driver.onvifClients = map[string]*OnvifClient{
				testDeviceName: client,
			}

			mockService.On("GetDeviceByName", testDeviceName).
				Return(createTestDevice(), nil)

			mockDevice.On("GetEndpointByRequestStruct", mock.Anything).Return(server.URL, nil)

			sendSoap := mockDevice.On("SendSoap", mock.Anything, mock.Anything)
			sendSoap.Run(func(args mock.Arguments) {
				resp, err := http.Post(server.URL, "application/soap+xml; charset=utf-8", strings.NewReader(args.String(1)))
				sendSoap.Return(resp, err)
			})

			actual, err := driver.HandleReadCommands(test.deviceName, test.protocols, test.reqs)
			if test.errorExpected {
				require.Error(t, err)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}

// TestUpdateDevice verifies proper updating of device information
func TestUpdateDevice(t *testing.T) {
	driver, mockService := createDriverWithMockService()
	tests := []struct {
		device  models.Device
		devInfo *device.GetDeviceInformationResponse

		expectedDevice           models.Device
		errorExpected            bool
		updateDeviceExpected     bool
		addDeviceExpected        bool
		removeDeviceExpected     bool
		removeDeviceFailExpected bool
	}{
		{
			device: contract.Device{
				Name: "testName",
			},
			updateDeviceExpected: true,
			devInfo: &device.GetDeviceInformationResponse{
				Manufacturer:    "Intel",
				Model:           "SimCamera",
				FirmwareVersion: "2.5a",
				SerialNumber:    "9a32410c",
				HardwareId:      "1.0",
			},
		},
		{
			removeDeviceExpected:     true,
			removeDeviceFailExpected: true,
			addDeviceExpected:        true,
			device: contract.Device{
				Name: "unknown_unknown_device",
				Protocols: map[string]models.ProtocolProperties{
					OnvifProtocol: map[string]string{
						EndpointRefAddress: "793dfb2-28b0-11ed-a261-0242ac120002",
					},
				}},
			devInfo: &device.GetDeviceInformationResponse{
				Manufacturer:    "Intel",
				Model:           "SimCamera",
				FirmwareVersion: "2.5a",
				SerialNumber:    "9a32410c",
				HardwareId:      "1.0",
			},
			expectedDevice: contract.Device{
				Name: "Intel-SimCamera-793dfb2-28b0-11ed-a261-0242ac120002",
				Protocols: map[string]models.ProtocolProperties{
					OnvifProtocol: map[string]string{
						EndpointRefAddress: "793dfb2-28b0-11ed-a261-0242ac120002",
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.device.Name, func(t *testing.T) {

			if test.removeDeviceExpected {
				if test.removeDeviceFailExpected {
					mockService.On("RemoveDeviceByName", test.device.Name).Return(errors.NewCommonEdgeX(errors.KindContractInvalid, "unit test error", nil)).Once()
				} else {
					mockService.On("RemoveDeviceByName", test.device.Name).Return(nil).Once()
				}
			}

			if test.updateDeviceExpected {
				mockService.On("UpdateDevice", test.device).Return(nil).Once()
			}

			if test.addDeviceExpected {
				mockService.On("AddDevice", test.expectedDevice).Return(test.expectedDevice.Name, nil).Once()
			}

			err := driver.updateDevice(test.device, test.devInfo)

			mockService.AssertExpectations(t)
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestDriver_RemoveDevice tests the different code flows of the RemoveDevice when called with an actual device
// versus when called with the control plane device.
func TestDriver_RemoveDevice(t *testing.T) {
	driver, mockService := createDriverWithMockService()
	driver.asynchCh = make(chan *sdkModel.AsyncValues, 1)
	driver.clientsMu = new(sync.RWMutex)
	driver.configMu = new(sync.RWMutex)
	driver.onvifClients = make(map[string]*OnvifClient)

	tests := []struct {
		name       string
		deviceName string
		wantErr    bool
	}{
		{
			name:       "control plane device",
			deviceName: "device-onvif-camera",
		},
		{
			name:       "regular onvif device",
			deviceName: "my-added-device",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			mockService.On("Name").Return(test.deviceName)

			err := driver.RemoveDevice(test.deviceName, map[string]models.ProtocolProperties{})
			if test.wantErr {
				require.Error(t, err)
			}
			mockService.AssertExpectations(t)
		})
	}
}
