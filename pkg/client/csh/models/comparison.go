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

// Comparison TODO - "comparison" does not sound apt as a name
//
// swagger:model Comparison
type Comparison struct {

	// result
	Result bool `json:"result,omitempty"`
}

// Validate validates this comparison
func (m *Comparison) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this comparison based on context it is used
func (m *Comparison) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Comparison) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Comparison) UnmarshalBinary(b []byte) error {
	var res Comparison
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
