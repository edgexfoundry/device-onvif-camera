// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	http "net/http"

	onvif "github.com/IOTechSystems/onvif"
	mock "github.com/stretchr/testify/mock"
)

// OnvifDevice is an autogenerated mock type for the OnvifDevice type
type OnvifDevice struct {
	mock.Mock
}

// CallMethod provides a mock function with given fields: method
func (_m *OnvifDevice) CallMethod(method interface{}) (*http.Response, error) {
	ret := _m.Called(method)

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}) (*http.Response, error)); ok {
		return rf(method)
	}
	if rf, ok := ret.Get(0).(func(interface{}) *http.Response); ok {
		r0 = rf(method)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}) error); ok {
		r1 = rf(method)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CallOnvifFunction provides a mock function with given fields: serviceName, functionName, data
func (_m *OnvifDevice) CallOnvifFunction(serviceName string, functionName string, data []byte) (interface{}, error) {
	ret := _m.Called(serviceName, functionName, data)

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, []byte) (interface{}, error)); ok {
		return rf(serviceName, functionName, data)
	}
	if rf, ok := ret.Get(0).(func(string, string, []byte) interface{}); ok {
		r0 = rf(serviceName, functionName, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, []byte) error); ok {
		r1 = rf(serviceName, functionName, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeviceInfo provides a mock function with given fields:
func (_m *OnvifDevice) GetDeviceInfo() onvif.DeviceInfo {
	ret := _m.Called()

	var r0 onvif.DeviceInfo
	if rf, ok := ret.Get(0).(func() onvif.DeviceInfo); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(onvif.DeviceInfo)
	}

	return r0
}

// GetDeviceParams provides a mock function with given fields:
func (_m *OnvifDevice) GetDeviceParams() onvif.DeviceParams {
	ret := _m.Called()

	var r0 onvif.DeviceParams
	if rf, ok := ret.Get(0).(func() onvif.DeviceParams); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(onvif.DeviceParams)
	}

	return r0
}

// GetEndpoint provides a mock function with given fields: name
func (_m *OnvifDevice) GetEndpoint(name string) string {
	ret := _m.Called(name)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetEndpointByRequestStruct provides a mock function with given fields: requestStruct
func (_m *OnvifDevice) GetEndpointByRequestStruct(requestStruct interface{}) (string, error) {
	ret := _m.Called(requestStruct)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}) (string, error)); ok {
		return rf(requestStruct)
	}
	if rf, ok := ret.Get(0).(func(interface{}) string); ok {
		r0 = rf(requestStruct)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(interface{}) error); ok {
		r1 = rf(requestStruct)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServices provides a mock function with given fields:
func (_m *OnvifDevice) GetServices() map[string]string {
	ret := _m.Called()

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// SendGetSnapshotRequest provides a mock function with given fields: url
func (_m *OnvifDevice) SendGetSnapshotRequest(url string) (*http.Response, error) {
	ret := _m.Called(url)

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*http.Response, error)); ok {
		return rf(url)
	}
	if rf, ok := ret.Get(0).(func(string) *http.Response); ok {
		r0 = rf(url)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendSoap provides a mock function with given fields: endpoint, xmlRequestBody
func (_m *OnvifDevice) SendSoap(endpoint string, xmlRequestBody string) (*http.Response, error) {
	ret := _m.Called(endpoint, xmlRequestBody)

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (*http.Response, error)); ok {
		return rf(endpoint, xmlRequestBody)
	}
	if rf, ok := ret.Get(0).(func(string, string) *http.Response); ok {
		r0 = rf(endpoint, xmlRequestBody)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(endpoint, xmlRequestBody)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewOnvifDevice interface {
	mock.TestingT
	Cleanup(func())
}

// NewOnvifDevice creates a new instance of OnvifDevice. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOnvifDevice(t mockConstructorTestingTNewOnvifDevice) *OnvifDevice {
	mock := &OnvifDevice{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
