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

package identity

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/openziti/edge-api/rest_model"
)

// RemoveIdentityMfaReader is a Reader for the RemoveIdentityMfa structure.
type RemoveIdentityMfaReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *RemoveIdentityMfaReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewRemoveIdentityMfaOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewRemoveIdentityMfaUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewRemoveIdentityMfaNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 429:
		result := NewRemoveIdentityMfaTooManyRequests()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 503:
		result := NewRemoveIdentityMfaServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewRemoveIdentityMfaOK creates a RemoveIdentityMfaOK with default headers values
func NewRemoveIdentityMfaOK() *RemoveIdentityMfaOK {
	return &RemoveIdentityMfaOK{}
}

/* RemoveIdentityMfaOK describes a response with status code 200, with default header values.

Base empty response
*/
type RemoveIdentityMfaOK struct {
	Payload *rest_model.Empty
}

func (o *RemoveIdentityMfaOK) Error() string {
	return fmt.Sprintf("[DELETE /identities/{id}/mfa][%d] removeIdentityMfaOK  %+v", 200, o.Payload)
}
func (o *RemoveIdentityMfaOK) GetPayload() *rest_model.Empty {
	return o.Payload
}

func (o *RemoveIdentityMfaOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.Empty)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRemoveIdentityMfaUnauthorized creates a RemoveIdentityMfaUnauthorized with default headers values
func NewRemoveIdentityMfaUnauthorized() *RemoveIdentityMfaUnauthorized {
	return &RemoveIdentityMfaUnauthorized{}
}

/* RemoveIdentityMfaUnauthorized describes a response with status code 401, with default header values.

The supplied session does not have the correct access rights to request this resource
*/
type RemoveIdentityMfaUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *RemoveIdentityMfaUnauthorized) Error() string {
	return fmt.Sprintf("[DELETE /identities/{id}/mfa][%d] removeIdentityMfaUnauthorized  %+v", 401, o.Payload)
}
func (o *RemoveIdentityMfaUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *RemoveIdentityMfaUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRemoveIdentityMfaNotFound creates a RemoveIdentityMfaNotFound with default headers values
func NewRemoveIdentityMfaNotFound() *RemoveIdentityMfaNotFound {
	return &RemoveIdentityMfaNotFound{}
}

/* RemoveIdentityMfaNotFound describes a response with status code 404, with default header values.

The requested resource does not exist
*/
type RemoveIdentityMfaNotFound struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *RemoveIdentityMfaNotFound) Error() string {
	return fmt.Sprintf("[DELETE /identities/{id}/mfa][%d] removeIdentityMfaNotFound  %+v", 404, o.Payload)
}
func (o *RemoveIdentityMfaNotFound) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *RemoveIdentityMfaNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRemoveIdentityMfaTooManyRequests creates a RemoveIdentityMfaTooManyRequests with default headers values
func NewRemoveIdentityMfaTooManyRequests() *RemoveIdentityMfaTooManyRequests {
	return &RemoveIdentityMfaTooManyRequests{}
}

/* RemoveIdentityMfaTooManyRequests describes a response with status code 429, with default header values.

The resource requested is rate limited and the rate limit has been exceeded
*/
type RemoveIdentityMfaTooManyRequests struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *RemoveIdentityMfaTooManyRequests) Error() string {
	return fmt.Sprintf("[DELETE /identities/{id}/mfa][%d] removeIdentityMfaTooManyRequests  %+v", 429, o.Payload)
}
func (o *RemoveIdentityMfaTooManyRequests) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *RemoveIdentityMfaTooManyRequests) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRemoveIdentityMfaServiceUnavailable creates a RemoveIdentityMfaServiceUnavailable with default headers values
func NewRemoveIdentityMfaServiceUnavailable() *RemoveIdentityMfaServiceUnavailable {
	return &RemoveIdentityMfaServiceUnavailable{}
}

/* RemoveIdentityMfaServiceUnavailable describes a response with status code 503, with default header values.

The request could not be completed due to the server being busy or in a temporarily bad state
*/
type RemoveIdentityMfaServiceUnavailable struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *RemoveIdentityMfaServiceUnavailable) Error() string {
	return fmt.Sprintf("[DELETE /identities/{id}/mfa][%d] removeIdentityMfaServiceUnavailable  %+v", 503, o.Payload)
}
func (o *RemoveIdentityMfaServiceUnavailable) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *RemoveIdentityMfaServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
