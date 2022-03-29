// Code generated by go-swagger; DO NOT EDIT.

// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new operations API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for operations API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientService is the interface for Client methods
type ClientService interface {
	GetConfig(params *GetConfigParams) (*GetConfigOK, error)

	PostAuthorizations(params *PostAuthorizationsParams) (*PostAuthorizationsOK, error)

	PostCompare(params *PostCompareParams) (*PostCompareOK, error)

	PostExtract(params *PostExtractParams) (*PostExtractOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  GetConfig Returns the Comparator's auto-generated configuration.

This configuration may be used for instance to configure a profile in the VC HTTP API for issuance of
Verifiable Credentials using the same DID and keys.

*/
func (a *Client) GetConfig(params *GetConfigParams) (*GetConfigOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetConfigParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "GetConfig",
		Method:             "GET",
		PathPattern:        "/config",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetConfigReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetConfigOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for GetConfig: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  PostAuthorizations Authorize a third party to perform a comparison on a Vault Server document.

Authorization to read the document is obtained at the Vault Server and pre-configured in the remote
Confidential Storage Hub, to be referenced during the actual comparison operation.

*/
func (a *Client) PostAuthorizations(params *PostAuthorizationsParams) (*PostAuthorizationsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostAuthorizationsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "PostAuthorizations",
		Method:             "POST",
		PathPattern:        "/authorizations",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostAuthorizationsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostAuthorizationsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for PostAuthorizations: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  PostCompare Execute a _remote_ comparison of the Confidential Storage documents fetched with the credentials provided.
This comparison is performed remotely by the Confidential Storage hub using the credentials.

The comparison's operator's type determines the type of comparison to be performed.

The result is always a boolean value.

*/
func (a *Client) PostCompare(params *PostCompareParams) (*PostCompareOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostCompareParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "PostCompare",
		Method:             "POST",
		PathPattern:        "/compare",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostCompareReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostCompareOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for PostCompare: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  PostExtract Extract the contents of one or more documents using the authorization tokens provided. The tokens originate
from authorizations granted at other Comparators.

*/
func (a *Client) PostExtract(params *PostExtractParams) (*PostExtractOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostExtractParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "PostExtract",
		Method:             "POST",
		PathPattern:        "/extract",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostExtractReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostExtractOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for PostExtract: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
