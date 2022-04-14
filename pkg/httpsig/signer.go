/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpsig

import (
	"crypto/ed25519"
	"fmt"
	"net/http"
	"time"

	httpsig "github.com/igor-pavlenko/httpsignatures-go"
	"github.com/trustbloc/edge-core/pkg/log"
)

var logger = log.New("httpsig")

const (
	dateHeader = "Date"
)

// DefaultGetSignerConfig returns the default configuration for signing HTTP GET requests.
func DefaultGetSignerConfig() SignerConfig {
	return SignerConfig{
		Headers: []string{"(request-target)", "Date"},
	}
}

// DefaultPostSignerConfig returns the default configuration for signing HTTP POST requests.
func DefaultPostSignerConfig() SignerConfig {
	return SignerConfig{
		Headers: []string{"(request-target)", "Date", "Digest"},
	}
}

// SignerConfig contains the configuration for signing HTTP requests.
type SignerConfig struct {
	Headers []string
}

type signer interface {
	Sign(secretKeyID string, r *http.Request) error
}

// Signer signs HTTP requests.
type Signer struct {
	SignerConfig
	signer func() signer
}

// NewSigner returns a new signer.
func NewSigner(cfg SignerConfig, privateKey ed25519.PrivateKey) *Signer {
	algo := NewSignerAlgorithm(privateKey)
	secretRetriever := &SecretRetriever{}

	return &Signer{
		SignerConfig: cfg,
		signer: func() signer {
			// Return a new instance for each signature since the HTTP signature
			// implementation is not thread safe.
			hs := httpsig.NewHTTPSignatures(secretRetriever)
			hs.SetDefaultSignatureHeaders(cfg.Headers)
			hs.SetSignatureHashAlgorithm(algo)

			return hs
		},
	}
}

// SignRequest signs an HTTP request.
func (s *Signer) SignRequest(pubKeyID string, req *http.Request) error {
	req.Header.Add(dateHeader, date())

	logger.Debugf("Signing request for %s. Public key ID [%s]. Headers: %s", req.RequestURI, pubKeyID, req.Header)

	if err := s.signer().Sign(pubKeyID, req); err != nil {
		return fmt.Errorf("sign request with public key ID [%s]: %w", pubKeyID, err)
	}

	logger.Debugf("Signed request. Headers: %s", req.Header)

	return nil
}

func date() string {
	return fmt.Sprintf("%s GMT", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05"))
}
