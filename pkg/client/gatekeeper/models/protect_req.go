// Code generated by go-swagger; DO NOT EDIT.

// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProtectReq protect req
//
// swagger:model ProtectReq
type ProtectReq struct {

	// policy
	Policy string `json:"policy,omitempty"`

	// target
	Target string `json:"target,omitempty"`
}

// Validate validates this protect req
func (m *ProtectReq) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this protect req based on context it is used
func (m *ProtectReq) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ProtectReq) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProtectReq) UnmarshalBinary(b []byte) error {
	var res ProtectReq
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}