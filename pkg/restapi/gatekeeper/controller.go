/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gatekeeper

import (
	"fmt"

	"github.com/trustbloc/ace/pkg/internal/common/support"
	"github.com/trustbloc/ace/pkg/restapi/gatekeeper/operation"
)

// New returns new controller instance.
func New(cfg *operation.Config) (*Controller, error) {
	ops, err := operation.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize operation: %w", err)
	}

	return &Controller{handlers: ops.GetRESTHandlers()}, nil
}

// Controller contains handlers for controller.
type Controller struct {
	handlers []support.Handler
}

// GetOperations returns all controller endpoints.
func (c *Controller) GetOperations() []support.Handler {
	return c.handlers
}
