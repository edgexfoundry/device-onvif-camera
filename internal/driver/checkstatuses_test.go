// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateDeviceStatus_update(t *testing.T) {
	driver, mockService := createDriverWithMockService()
	mockService.On("GetDeviceByName", testDeviceName).
		Return(createTestDevice(), nil).Once()
	mockService.On("UpdateDevice", mock.AnythingOfType("models.Device")).
		Return(nil).Once()

	changed, err := driver.updateDeviceStatus(testDeviceName, UpWithAuth)
	mockService.AssertExpectations(t)
	require.NoError(t, err)
	assert.True(t, changed)
}

func TestUpdateDeviceStatus_noDevice(t *testing.T) {
	driver, mockService := createDriverWithMockService()
	mockService.On("GetDeviceByName", testDeviceName).
		Return(models.Device{}, errors.New("error")).Once()

	_, err := driver.updateDeviceStatus(testDeviceName, UpWithAuth)
	mockService.AssertExpectations(t)
	require.Error(t, err)
}

func TestUpdateDeviceStatus_noUpdate(t *testing.T) {
	driver, mockService := createDriverWithMockService()
	mockService.On("GetDeviceByName", testDeviceName).
		Return(createTestDevice(), nil).Once()

	changed, err := driver.updateDeviceStatus(testDeviceName, Unreachable)
	mockService.AssertExpectations(t)
	require.NoError(t, err)
	assert.False(t, changed)
}

// func TestDriver_TestConnectionMethod(t *testing.T) {
// 	driver, _ := createDriverWithMockService()
// 	driver.clientsMu = new(sync.RWMutex)

// 	tests := []struct {
// 		name   string
// 		device models.Device
// 		// protocols     map[string]models.ProtocolProperties
// 		reqs          []sdkModel.CommandRequest
// 		resp          string
// 		data          string
// 		expected      string
// 		errorExpected bool
// 	}{
// 		{
// 			name:   "simple read of DeviceInformation",
// 			device: createTestDevice(),
// 			reqs: []sdkModel.CommandRequest{
// 				{
// 					DeviceResourceName: "DeviceInformation",
// 					Attributes: map[string]interface{}{
// 						getFunction: "GetDeviceInformation",
// 						"service":   onvif.DeviceWebService,
// 					},
// 					Type: "Object",
// 				}},
// 			resp: `<?xml version="1.0" encoding="UTF-8"?>
// <Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
//   <Header />
//   <Body>
//     <Content>
//       <Manufacturer>Intel</Manufacturer>
//       <Model>SimCamera</Model>
//       <FirmwareVersion>2.4a</FirmwareVersion>
//       <SerialNumber>46d1ab8d</SerialNumber>
//       <HardwareId>1.0</HardwareId>
//     </Content>
//   </Body>
// </Envelope>`,
// 			expected: Unreachable,
// 		},
// 	}

// 	for _, test := range tests {
// 		test := test
// 		t.Run(test.name, func(t *testing.T) {
// 			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
// 				_, err := writer.Write([]byte(test.resp))
// 				assert.NoError(t, err)
// 			}))
// 			defer server.Close()

// 			client, mockDevice := createOnvifClientWithMockDevice(driver, testDeviceName)
// 			driver.onvifClients = map[string]*OnvifClient{
// 				testDeviceName: client,
// 			}

// 			mockDevice.On("callOnvifFunction", mock.Anything).Return(server.URL, nil)

// 			sendSoap := mockDevice.On("SendSoap", mock.Anything, mock.Anything)
// 			sendSoap.Run(func(args mock.Arguments) {
// 				resp, err := http.Post(server.URL, "application/soap+xml; charset=utf-8", strings.NewReader(args.String(1)))
// 				sendSoap.Return(resp, err)
// 			})
// 			substrings := strings.Split(server.URL, ":")
// 			test.device.Protocols[OnvifProtocol][Address] = substrings[1][2:len(substrings[1])]
// 			test.device.Protocols[OnvifProtocol][Port] = substrings[2]

// 			actual := driver.testConnectionMethods(test.device)
// 			assert.Equal(t, test.expected, actual)
// 		})
// 	}
// }

func TestDriver_TCPProbe(t *testing.T) {
	driver, _ := createDriverWithMockService()
	driver.clientsMu = new(sync.RWMutex)
	driver.config = &ServiceConfig{
		AppCustom: CustomConfig{
			ProbeTimeoutMillis: 2000,
		},
	}

	tests := []struct {
		name         string
		device       models.Device
		deviceExists bool
		expected     bool
	}{
		{
			name: "properConnection",
			device: models.Device{
				Protocols: map[string]models.ProtocolProperties{
					OnvifProtocol: {},
				},
			},
			deviceExists: true,
			expected:     true,
		},
		{
			name: "emptyAddress",
			device: models.Device{
				Protocols: map[string]models.ProtocolProperties{
					OnvifProtocol: {
						Address: "",
						Port:    "",
					},
				},
			},
			expected: false,
		},
		{
			name: "emptyProtocols",
			device: models.Device{
				Protocols: map[string]models.ProtocolProperties{},
			},
			expected: false,
		},
		{
			name: "noDevice",
			device: models.Device{
				Protocols: map[string]models.ProtocolProperties{
					OnvifProtocol: {
						Address: "1.1.1.1",
						Port:    "1",
					},
				},
			},
			deviceExists: false,
			expected:     false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.deviceExists {
				server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					_, err := writer.Write([]byte(nil))
					assert.NoError(t, err)
				}))
				defer server.Close()
				substrings := strings.Split(server.URL, ":")
				test.device.Protocols[OnvifProtocol][Address] = substrings[1][2:len(substrings[1])]
				test.device.Protocols[OnvifProtocol][Port] = substrings[2]
			}

			actual := driver.tcpProbe(test.device)
			assert.Equal(t, test.expected, actual)
		})
	}
}
