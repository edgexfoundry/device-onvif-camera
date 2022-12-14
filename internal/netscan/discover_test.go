// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package netscan

import (
	"context"
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// startServerWithResponse starts a new httptest.Server, and returns the server plus the port number.
// NOTE: defer server.Close() should be called immediately after calling this function.
func startServerWithResponse(t *testing.T, response string) (server *httptest.Server, port string) {
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(response))
		assert.NoError(t, err)
	}))

	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	return server, u.Port()
}

func TestAutoDiscover_EmptyOrInvalidSubnets(t *testing.T) {
	server, port := startServerWithResponse(t, "Hello World!")
	defer server.Close()

	tests := []struct {
		name    string
		subnets []string
	}{
		{
			name:    "nil subnets",
			subnets: nil,
		},
		{
			name:    "empty subnets",
			subnets: []string{"", ""},
		},
		{
			name:    "ipv6 subnets",
			subnets: []string{"2001:4860:4860::8888/32"},
		},
		{
			name:    "invalid cidr",
			subnets: []string{"1.1/2"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			params := Params{
				Subnets:         test.subnets,
				AsyncLimit:      100,
				Timeout:         time.Duration(100) * time.Millisecond,
				ScanPorts:       []string{port},
				Logger:          logger.NewMockClient(),
				NetworkProtocol: NetworkTCP,
			}

			ctx, cancel := context.WithTimeout(context.Background(),
				time.Duration(5)*time.Second)
			defer cancel()

			mockProtocol := MockProtocolSpecificDiscovery{}

			result := AutoDiscover(ctx, &mockProtocol, params)
			assert.Empty(t, result)
		})
	}
}

func TestAutoDiscover(t *testing.T) {
	server, port := startServerWithResponse(t, "Hello World!")
	defer server.Close()

	params := Params{
		Subnets:         []string{"127.0.0.1/32"},
		AsyncLimit:      100,
		Timeout:         time.Duration(100) * time.Millisecond,
		ScanPorts:       []string{port},
		Logger:          logger.NewMockClient(),
		NetworkProtocol: NetworkTCP,
	}

	testDeviceName := "test-discovered-device"

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(5)*time.Second)
	defer cancel()

	mockProtocol := MockProtocolSpecificDiscovery{}

	probeFilter := mockProtocol.On("ProbeFilter", mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Once()
	probeFilter.Run(func(args mock.Arguments) {
		// return the ports as is
		probeFilter.Return(args.Get(1).([]string))
	})

	connDialed := mockProtocol.On("OnConnectionDialed", mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.Anything, mock.Anything).Once()
	connDialed.Run(func(args mock.Arguments) {
		connDialed.Return([]ProbeResult{{
			Host: args.String(0),
			Port: args.String(1),
			Data: "",
		}}, nil)
	})

	convertResult := mockProtocol.On("ConvertProbeResult", mock.Anything, mock.Anything).Once()
	convertResult.Run(func(args mock.Arguments) {
		convertResult.Return(models.DiscoveredDevice{
			Name: testDeviceName,
			Protocols: map[string]contract.ProtocolProperties{
				"tcp": {
					"Address": args.Get(0).(ProbeResult).Host,
					"Port":    args.Get(0).(ProbeResult).Port,
				},
			},
			Description: "Example discovered device",
			Labels:      []string{},
		}, nil)
	})

	result := AutoDiscover(ctx, &mockProtocol, params)
	mockProtocol.AssertExpectations(t)
	assert.NotEmpty(t, result)
	assert.Equal(t, testDeviceName, result[0].Name)
}

func TestAutoDiscover_MultiPort(t *testing.T) {
	server1, port1 := startServerWithResponse(t, "Hello World from server 1!")
	defer server1.Close()

	server2, port2 := startServerWithResponse(t, "Hello World from server 2!")
	defer server2.Close()

	server3, port3 := startServerWithResponse(t, "Hello World from server 3!")
	defer server3.Close()

	params := Params{
		Subnets:         []string{"127.0.0.1/32"},
		AsyncLimit:      100,
		Timeout:         time.Duration(100) * time.Millisecond,
		ScanPorts:       []string{port1, port2, port3},
		Logger:          logger.NewMockClient(),
		NetworkProtocol: NetworkTCP,
	}

	testDeviceName := "test-discovered-device"

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(5)*time.Second)
	defer cancel()

	mockProtocol := MockProtocolSpecificDiscovery{}

	probeFilter := mockProtocol.On("ProbeFilter", mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Once()
	probeFilter.Run(func(args mock.Arguments) {
		// return the ports as is
		probeFilter.Return(args.Get(1).([]string))
	})

	connDialed := mockProtocol.On("OnConnectionDialed", mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.Anything, mock.Anything).Times(3)
	connDialed.Run(func(args mock.Arguments) {
		connDialed.Return([]ProbeResult{{
			Host: args.String(0),
			Port: args.String(1),
			Data: "",
		}}, nil)
	})

	convertResult := mockProtocol.On("ConvertProbeResult", mock.Anything, mock.Anything).Times(3)
	convertResult.Run(func(args mock.Arguments) {
		convertResult.Return(models.DiscoveredDevice{
			Name: testDeviceName,
			Protocols: map[string]contract.ProtocolProperties{
				"tcp": {
					"Address": args.Get(0).(ProbeResult).Host,
					"Port":    args.Get(0).(ProbeResult).Port,
				},
			},
			Description: "Example discovered device",
			Labels:      []string{},
		}, nil)
	})

	results := AutoDiscover(ctx, &mockProtocol, params)
	mockProtocol.AssertExpectations(t)
	assert.NotEmpty(t, results)
	assert.Len(t, results, 3)
	for _, result := range results {
		assert.Contains(t, []string{port1, port2, port3}, result.Protocols["tcp"]["Port"])
	}
}
