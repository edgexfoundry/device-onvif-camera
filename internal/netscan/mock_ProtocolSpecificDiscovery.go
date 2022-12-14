// Code generated by mockery v2.13.1. DO NOT EDIT.

package netscan

import (
	net "net"

	models "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	mock "github.com/stretchr/testify/mock"
)

// MockProtocolSpecificDiscovery is an autogenerated mock type for the ProtocolSpecificDiscovery type
type MockProtocolSpecificDiscovery struct {
	mock.Mock
}

// ConvertProbeResult provides a mock function with given fields: probeResult, params
func (_m *MockProtocolSpecificDiscovery) ConvertProbeResult(probeResult ProbeResult, params Params) (models.DiscoveredDevice, error) {
	ret := _m.Called(probeResult, params)

	var r0 models.DiscoveredDevice
	if rf, ok := ret.Get(0).(func(ProbeResult, Params) models.DiscoveredDevice); ok {
		r0 = rf(probeResult, params)
	} else {
		r0 = ret.Get(0).(models.DiscoveredDevice)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ProbeResult, Params) error); ok {
		r1 = rf(probeResult, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OnConnectionDialed provides a mock function with given fields: host, port, conn, params
func (_m *MockProtocolSpecificDiscovery) OnConnectionDialed(host string, port string, conn net.Conn, params Params) ([]ProbeResult, error) {
	ret := _m.Called(host, port, conn, params)

	var r0 []ProbeResult
	if rf, ok := ret.Get(0).(func(string, string, net.Conn, Params) []ProbeResult); ok {
		r0 = rf(host, port, conn, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]ProbeResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, net.Conn, Params) error); ok {
		r1 = rf(host, port, conn, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProbeFilter provides a mock function with given fields: host, ports
func (_m *MockProtocolSpecificDiscovery) ProbeFilter(host string, ports []string) []string {
	ret := _m.Called(host, ports)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string, []string) []string); ok {
		r0 = rf(host, ports)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

type mockConstructorTestingTNewMockProtocolSpecificDiscovery interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockProtocolSpecificDiscovery creates a new instance of MockProtocolSpecificDiscovery. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockProtocolSpecificDiscovery(t mockConstructorTestingTNewMockProtocolSpecificDiscovery) *MockProtocolSpecificDiscovery {
	mock := &MockProtocolSpecificDiscovery{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}