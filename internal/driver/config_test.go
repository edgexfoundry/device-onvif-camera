package driver

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFromRaw(t *testing.T) {
	expectedConfig := &ServiceConfig{
		AppCustom: CustomConfig{
			RequestTimeout:             5,
			DefaultSecretName:          "default_secret",
			DiscoveryEthernetInterface: "eth0",
			BaseNotificationURL:        "http://localhost:59984",
			DiscoveryMode:              "netscan",
			DiscoverySubnets:           "127.0.0.1/32,127.0.1.1/32",
			ProbeAsyncLimit:            50,
			ProbeTimeoutMillis:         1000,
			MaxDiscoverDurationSeconds: 5,
			ProvisionWatcherDir:        "res/provision_watchers",
		},
	}
	testCases := []struct {
		Name      string
		rawConfig interface{}
		isValid   bool
	}{
		{
			Name:      "valid",
			isValid:   true,
			rawConfig: expectedConfig,
		},
		{
			Name:      "not valid",
			isValid:   false,
			rawConfig: expectedConfig.AppCustom,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			actualConfig := ServiceConfig{}

			ok := actualConfig.UpdateFromRaw(testCase.rawConfig)

			assert.Equal(t, testCase.isValid, ok)
			if testCase.isValid {
				assert.Equal(t, expectedConfig, &actualConfig)
			}
		})
	}
}

// TestGetCameraXAddr verifies the parsing of the camera XAddr
func TestGetCameraXAddr(t *testing.T) {

	tests := []struct {
		input         map[string]models.ProtocolProperties
		expected      string
		errorExpected bool
	}{
		{
			input: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					Address: "localhost",
					Port:    "80",
				},
			},
			expected: "localhost:80",
		},
		{
			input: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					Address: "localhost",
				},
			},
			expected: "localhost",
		},
		{
			input: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					CustomMetadata: "custommetadata",
				},
			},
			errorExpected: true,
		},
		{
			input: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					Address: "",
					Port:    "",
				},
			},
			errorExpected: true,
		},
		{
			input:         nil,
			errorExpected: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.expected, func(t *testing.T) {
			actual, err := GetCameraXAddr(test.input)

			if test.errorExpected {
				require.Error(t, err)
				return
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}
