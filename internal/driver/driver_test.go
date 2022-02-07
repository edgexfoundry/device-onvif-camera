// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/base64"
	"fmt"
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
