package driver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateFromRaw(t *testing.T) {
	expectedConfig := &ServiceConfig{
		AppCustom: CustomConfig{
			RequestTimeout:             5,
			DefaultSecretPath:          "default_secret",
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
