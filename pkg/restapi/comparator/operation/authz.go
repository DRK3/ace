/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operation

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/jsonld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	"github.com/square/go-jose/v3"
	"github.com/trustbloc/edge-core/pkg/zcapld"

	"github.com/trustbloc/ace/pkg/client/csh/client/operations"
	cshclientmodels "github.com/trustbloc/ace/pkg/client/csh/models"
	"github.com/trustbloc/ace/pkg/restapi/comparator/operation/models"
)

// HandleAuthz handles a CreateAuthzReq.
func (o *Operation) HandleAuthz(w http.ResponseWriter, authz *models.Authorization) { //nolint: funlen
	docMeta, err := o.vaultClient.GetDocMetaData(authz.Scope.VaultID, *authz.Scope.DocID)
	if err != nil {
		respondErrorf(w, http.StatusInternalServerError, "failed to get doc meta: %s", err.Error())

		return
	}

	kmsURL, err := url.Parse(docMeta.EncKeyURI)
	if err != nil {
		respondErrorf(w, http.StatusInternalServerError, "failed to parse enc key uri: %s", err.Error())

		return
	}

	edvURL, err := url.Parse(docMeta.URI)
	if err != nil {
		respondErrorf(w, http.StatusInternalServerError, "failed to parse doc uri: %s", err.Error())

		return
	}

	parts := strings.Split(docMeta.URI, "/")

	vaultID := parts[len(parts)-3]
	docID := parts[len(parts)-1]

	response, err := o.cshClient.PostHubstoreProfilesProfileIDQueries(
		operations.NewPostHubstoreProfilesProfileIDQueriesParams().
			WithTimeout(requestTimeout).
			WithProfileID(o.cshProfile.ID).
			WithRequest(&cshclientmodels.DocQuery{
				VaultID: &vaultID,
				DocID:   &docID,
				Path:    authz.Scope.DocAttrPath,
				UpstreamAuth: &cshclientmodels.DocQueryAO1UpstreamAuth{
					Edv: &cshclientmodels.UpstreamAuthorization{
						BaseURL: fmt.Sprintf("%s://%s/%s", edvURL.Scheme, edvURL.Host, parts[3]),
						Zcap:    authz.Scope.AuthTokens.Edv,
					},
					Kms: &cshclientmodels.UpstreamAuthorization{
						BaseURL: fmt.Sprintf("%s://%s", kmsURL.Scheme, kmsURL.Host),
						Zcap:    authz.Scope.AuthTokens.Kms,
					},
				},
			}))
	if err != nil {
		respondErrorf(w, http.StatusInternalServerError, "failed to create query: %s", err.Error())

		return
	}

	// TODO - encode docPathAttr in zcap token
	// deriving a child zcap for csh
	zcap, err := o.driveZCAPForCSH(*authz.RequestingParty, response.Location,
		authz.Scope.Caveats())
	if err != nil {
		respondErrorf(w, http.StatusInternalServerError, "failed to drive child zcap from csh zcap: %s", err.Error())

		return
	}

	authToken, err := zcapld.CompressZCAP(zcap)
	if err != nil {
		respondErrorf(w, http.StatusInternalServerError, "failed to compress zcap: %s", err.Error())

		return
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	respond(w, http.StatusOK, headers, models.Authorization{
		RequestingParty: authz.RequestingParty,
		AuthToken:       authToken,
	})
}

func (o *Operation) driveZCAPForCSH(invokerDID, queryIDPath string,
	caveats []models.Caveat) (*zcapld.Capability, error) {
	cshZCAP, err := zcapld.DecompressZCAP(o.cshProfile.Zcap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSH profile zcap: %w", err)
	}

	keyID, key, err := getKey(o.comparatorConfig)
	if err != nil {
		return nil, err
	}

	return zcapld.NewCapability(&zcapld.Signer{
		SignatureSuite:     ed25519signature2018.New(suite.WithSigner(&ed25519Signer{key: key})),
		SuiteType:          ed25519signature2018.SignatureType,
		VerificationMethod: fmt.Sprintf("%s#%s", *o.comparatorConfig.Did, keyID),
		ProcessorOpts:      []jsonld.ProcessorOpts{jsonld.WithDocumentLoader(o.documentLoader)},
	}, zcapld.WithParent(cshZCAP.ID), zcapld.WithInvoker(invokerDID),
		zcapld.WithAllowedActions("reference"),
		zcapld.WithCaveats(toZCaveats(caveats)...),
		zcapld.WithInvocationTarget(queryIDPath, "urn:confidentialstoragehub:query"),
		zcapld.WithCapabilityChain(cshZCAP.ID),
	)
}

func getKey(comparatorConfig *models.Config) (string, ed25519.PrivateKey, error) {
	keys, ok := comparatorConfig.Key.([]interface{})
	if !ok {
		return "", nil, fmt.Errorf("key is not array")
	}

	keyBytes, err := json.Marshal(keys[0])
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	jwk := jose.JSONWebKey{}
	if errUnmarshalJSON := jwk.UnmarshalJSON(keyBytes); errUnmarshalJSON != nil {
		return "", nil, fmt.Errorf("failed to unmarshal key to jwk: %w", errUnmarshalJSON)
	}

	k, ok := jwk.Key.(ed25519.PrivateKey)
	if !ok {
		return "", nil, fmt.Errorf("key is not ed25519")
	}

	return jwk.KeyID, k, nil
}

type ed25519Signer struct {
	key ed25519.PrivateKey
}

func (s *ed25519Signer) Sign(data []byte) ([]byte, error) {
	return ed25519.Sign(s.key, data), nil
}

func (s *ed25519Signer) Alg() string {
	return ""
}

func toZCaveats(caveats []models.Caveat) []zcapld.Caveat {
	zCaveats := make([]zcapld.Caveat, len(caveats))

	for i, caveat := range caveats {
		switch t := caveat.(type) { //nolint: gocritic
		case *models.ExpiryCaveat:
			zCaveats[i] = zcapld.Caveat{
				Type:     t.Type(),
				Duration: uint64(t.Duration),
			}
		}
	}

	return zCaveats
}
