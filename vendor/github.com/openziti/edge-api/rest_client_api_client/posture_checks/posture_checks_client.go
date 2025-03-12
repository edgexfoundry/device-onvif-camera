// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package posture_checks

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new posture checks API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for posture checks API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	CreatePostureResponse(params *CreatePostureResponseParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePostureResponseCreated, error)

	CreatePostureResponseBulk(params *CreatePostureResponseBulkParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePostureResponseBulkOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  CreatePostureResponse submits a posture response to a posture query

  Submits posture responses
*/
func (a *Client) CreatePostureResponse(params *CreatePostureResponseParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePostureResponseCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreatePostureResponseParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createPostureResponse",
		Method:             "POST",
		PathPattern:        "/posture-response",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &CreatePostureResponseReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreatePostureResponseCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createPostureResponse: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  CreatePostureResponseBulk submits multiple posture responses

  Submits posture responses
*/
func (a *Client) CreatePostureResponseBulk(params *CreatePostureResponseBulkParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePostureResponseBulkOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreatePostureResponseBulkParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createPostureResponseBulk",
		Method:             "POST",
		PathPattern:        "/posture-response-bulk",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &CreatePostureResponseBulkReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreatePostureResponseBulkOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createPostureResponseBulk: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
