// Code generated by go-swagger; DO NOT EDIT.

// Copyright Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// BPFMapStatus BPF map status
//
// +k8s:deepcopy-gen=true
//
// swagger:model BPFMapStatus
type BPFMapStatus struct {

	// Ratio of total system memory to use for dynamic sizing of BPF maps
	DynamicSizeRatio float64 `json:"dynamic-size-ratio,omitempty"`

	// BPF maps
	Maps []*BPFMapProperties `json:"maps"`
}

// Validate validates this b p f map status
func (m *BPFMapStatus) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateMaps(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *BPFMapStatus) validateMaps(formats strfmt.Registry) error {
	if swag.IsZero(m.Maps) { // not required
		return nil
	}

	for i := 0; i < len(m.Maps); i++ {
		if swag.IsZero(m.Maps[i]) { // not required
			continue
		}

		if m.Maps[i] != nil {
			if err := m.Maps[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("maps" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("maps" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this b p f map status based on the context it is used
func (m *BPFMapStatus) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateMaps(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *BPFMapStatus) contextValidateMaps(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Maps); i++ {

		if m.Maps[i] != nil {
			if err := m.Maps[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("maps" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("maps" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *BPFMapStatus) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *BPFMapStatus) UnmarshalBinary(b []byte) error {
	var res BPFMapStatus
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}