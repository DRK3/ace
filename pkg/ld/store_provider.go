/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ld

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/store/ld"
	"github.com/hyperledger/aries-framework-go/spi/storage"
)

// StoreProvider provides stores for JSON-LD contexts and remote providers.
type StoreProvider struct {
	ContextStore        ld.ContextStore
	RemoteProviderStore ld.RemoteProviderStore
}

// NewStoreProvider returns a new instance of StoreProvider.
func NewStoreProvider(storageProvider storage.Provider) (*StoreProvider, error) {
	contextStore, err := ld.NewContextStore(storageProvider)
	if err != nil {
		return nil, fmt.Errorf("create JSON-LD context store: %w", err)
	}

	remoteProviderStore, err := ld.NewRemoteProviderStore(storageProvider)
	if err != nil {
		return nil, fmt.Errorf("create remote provider store: %w", err)
	}

	return &StoreProvider{
		ContextStore:        contextStore,
		RemoteProviderStore: remoteProviderStore,
	}, nil
}

// JSONLDContextStore returns JSON-LD context store.
func (p *StoreProvider) JSONLDContextStore() ld.ContextStore { //nolint:ireturn
	return p.ContextStore
}

// JSONLDRemoteProviderStore returns JSON-LD remote provider store.
func (p *StoreProvider) JSONLDRemoteProviderStore() ld.RemoteProviderStore { //nolint:ireturn
	return p.RemoteProviderStore
}
