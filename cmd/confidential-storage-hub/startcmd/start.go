/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/orb"
	ldrest "github.com/hyperledger/aries-framework-go/pkg/controller/rest/ld"
	"github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	webcrypto "github.com/hyperledger/aries-framework-go/pkg/crypto/webkms"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/webkms"
	ldsvc "github.com/hyperledger/aries-framework-go/pkg/ld"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/noop"
	"github.com/hyperledger/aries-framework-go/pkg/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
	ariesstorage "github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/trustbloc/edge-core/pkg/log"
	cmdutils "github.com/trustbloc/edge-core/pkg/utils/cmd"
	tlsutils "github.com/trustbloc/edge-core/pkg/utils/tls"
	edv "github.com/trustbloc/edv/pkg/client"
	"github.com/trustbloc/edv/pkg/restapi/models"

	"github.com/trustbloc/ace/cmd/common"
	"github.com/trustbloc/ace/pkg/client/vault"
	"github.com/trustbloc/ace/pkg/did"
	crypto2 "github.com/trustbloc/ace/pkg/key"
	"github.com/trustbloc/ace/pkg/ld"
	"github.com/trustbloc/ace/pkg/restapi/csh"
	"github.com/trustbloc/ace/pkg/restapi/csh/operation"
	zcapld2 "github.com/trustbloc/ace/pkg/restapi/csh/operation/zcapld"
	"github.com/trustbloc/ace/pkg/restapi/healthcheck"
)

const (
	hostURLFlagName      = "host-url"
	hostURLFlagShorthand = "u"
	hostURLFlagUsage     = "Host URL to run the confidential storage hub instance on. Format: HostName:Port."
	hostURLEnvKey        = "CSH_HOST_URL"

	baseURLFlagName  = "base-url"
	baseURLEnvKey    = "BASE_URL"
	baseURLFlagUsage = "Optional. Base URL on which the CSH service is exposed to clients. Defaults to `host-url`."

	tlsSystemCertPoolFlagName  = "tls-systemcertpool"
	tlsSystemCertPoolFlagUsage = "Use system certificate pool." +
		" Possible values [true] [false]. Defaults to false if not set." +
		" Alternatively, this can be set with the following environment variable: " + tlsSystemCertPoolEnvKey
	tlsSystemCertPoolEnvKey = "CSH_TLS_SYSTEMCERTPOOL"

	tlsCACertsFlagName  = "tls-cacerts"
	tlsCACertsFlagUsage = "Comma-Separated list of ca certs path." +
		" Alternatively, this can be set with the following environment variable: " + tlsCACertsEnvKey
	tlsCACertsEnvKey = "CSH_TLS_CACERTS"

	tlsServeCertPathFlagName  = "tls-serve-cert"
	tlsServeCertPathFlagUsage = "Path to the server certificate to use when serving HTTPS." +
		" Alternatively, this can be set with the following environment variable: " + tlsServeCertPathEnvKey
	tlsServeCertPathEnvKey = "CSH_TLS_SERVE_CERT"

	tlsServeKeyPathFlagName  = "tls-serve-key"
	tlsServeKeyPathFlagUsage = "Path to the private key to use when serving HTTPS." +
		" Alternatively, this can be set with the following environment variable: " + tlsServeKeyPathFlagEnvKey
	tlsServeKeyPathFlagEnvKey = "CSH_TLS_SERVE_KEY"

	didDomainFlagName  = "trustbloc-did-domain"
	didDomainFlagUsage = "Optional. URL to the did:trustbloc consortium's domain." +
		" Alternatively, this can be set with the following environment variable: " + didDomainEnvKey
	didDomainEnvKey = "TRUSTBLOC_DID_DOMAIN"

	identityDIDMethodFlagName  = "identity-did-method"
	identityDIDMethodFlagUsage = "DID method to use to create the CSH's identity. Valid values are [key, trustbloc]." +
		" Defaults to 'key'." +
		" Alternatively, this can be set with the following environment variable: " + identityDIDMethodEnvKey
	identityDIDMethodEnvKey = "IDENTITY_DID_METHOD"

	didAnchorOriginFlagName  = "did-anchor-origin"
	didAnchorOriginEnvKey    = "CSH_DID_ANCHOR_ORIGIN"
	didAnchorOriginFlagUsage = "DID anchor origin." +
		" Alternatively, this can be set with the following environment variable: " + didAnchorOriginEnvKey

	requestTokensFlagName  = "request-tokens"
	requestTokensEnvKey    = "CSH_REQUEST_TOKENS" //nolint: gosec
	requestTokensFlagUsage = "Tokens used for http request " +
		" Alternatively, this can be set with the following environment variable: " + requestTokensEnvKey

	splitRequestTokenLength = 2
)

var logger = log.New("confidential-storage-hub/start")

type serviceParameters struct {
	host              string
	baseURL           string
	tlsParams         *tlsParameters
	dbParams          *common.DBParameters
	trustblocDomain   string
	identityDIDMethod string
	didAnchorOrigin   string
	requestTokens     map[string]string
}

type tlsParameters struct {
	systemCertPool bool
	serveCertPath  string
	serveKeyPath   string
	tlsConfig      *tls.Config
}

type server interface {
	ListenAndServe(host string, certFile, keyFile string, router http.Handler) error
}

// HTTPServer represents an actual HTTP server implementation.
type HTTPServer struct{}

// ListenAndServe starts the server using the standard Go HTTP server implementation.
func (s *HTTPServer) ListenAndServe(host, certFile, keyFile string, router http.Handler) error {
	if certFile == "" || keyFile == "" {
		return http.ListenAndServe(host, router)
	}

	return http.ListenAndServeTLS(host, certFile, keyFile, router)
}

// GetStartCmd returns the Cobra start command.
func GetStartCmd(srv server) *cobra.Command {
	cmd := createStartCmd(srv)

	createFlags(cmd)

	return cmd
}

func createStartCmd(srv server) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts a confidential-storage-hub server",
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := getParameters(cmd)
			if err != nil {
				return err
			}

			return startService(params, srv)
		},
	}
}

func getParameters(cmd *cobra.Command) (*serviceParameters, error) {
	host, err := cmdutils.GetUserSetVarFromString(cmd, hostURLFlagName, hostURLEnvKey, false)
	if err != nil {
		return nil, err
	}

	baseURL, err := cmdutils.GetUserSetVarFromString(cmd, baseURLFlagName, baseURLEnvKey, true)
	if err != nil {
		return nil, err
	}

	tlsParams, err := getTLS(cmd)
	if err != nil {
		return nil, err
	}

	dbParams, err := common.DBParams(cmd)
	if err != nil {
		return nil, err
	}

	trustblocDomain, err := cmdutils.GetUserSetVarFromString(cmd, didDomainFlagName, didDomainEnvKey, true)
	if err != nil {
		return nil, err
	}

	identityDIDMethod, err := cmdutils.GetUserSetVarFromString(
		cmd, identityDIDMethodFlagName, identityDIDMethodEnvKey, true)
	if err != nil {
		return nil, err
	}

	didAnchorOrigin := cmdutils.GetUserSetOptionalVarFromString(cmd, didAnchorOriginFlagName, didAnchorOriginEnvKey)

	if identityDIDMethod == "" {
		identityDIDMethod = "key"
	}

	requestTokens := getRequestTokens(cmd)

	return &serviceParameters{
		host:              host,
		tlsParams:         tlsParams,
		dbParams:          dbParams,
		baseURL:           baseURL,
		trustblocDomain:   trustblocDomain,
		identityDIDMethod: identityDIDMethod,
		didAnchorOrigin:   didAnchorOrigin,
		requestTokens:     requestTokens,
	}, err
}

func createFlags(cmd *cobra.Command) {
	common.Flags(cmd)
	cmd.Flags().StringP(hostURLFlagName, hostURLFlagShorthand, "", hostURLFlagUsage)
	cmd.Flags().StringP(baseURLFlagName, "", "", baseURLFlagUsage)
	cmd.Flags().StringP(tlsSystemCertPoolFlagName, "", "", tlsSystemCertPoolFlagUsage)
	cmd.Flags().StringArrayP(tlsCACertsFlagName, "", []string{}, tlsCACertsFlagUsage)
	cmd.Flags().StringP(tlsServeCertPathFlagName, "", "", tlsServeCertPathFlagUsage)
	cmd.Flags().StringP(tlsServeKeyPathFlagName, "", "", tlsServeKeyPathFlagUsage)
	cmd.Flags().StringP(didDomainFlagName, "", "", didDomainFlagUsage)
	cmd.Flags().StringP(identityDIDMethodFlagName, "", "", identityDIDMethodFlagUsage)
	cmd.Flags().StringP(didAnchorOriginFlagName, "", "", didAnchorOriginFlagUsage)
	cmd.Flags().StringArrayP(requestTokensFlagName, "", []string{}, requestTokensFlagUsage)
}

func getTLS(cmd *cobra.Command) (*tlsParameters, error) {
	tlsSystemCertPoolString := cmdutils.GetUserSetOptionalVarFromString(cmd, tlsSystemCertPoolFlagName,
		tlsSystemCertPoolEnvKey)

	tlsSystemCertPool := false

	if tlsSystemCertPoolString != "" {
		var err error

		tlsSystemCertPool, err = strconv.ParseBool(tlsSystemCertPoolString)
		if err != nil {
			return nil, err
		}
	}

	tlsCACerts := cmdutils.GetUserSetOptionalVarFromArrayString(cmd, tlsCACertsFlagName, tlsCACertsEnvKey)

	rootCAs, err := tlsutils.GetCertPool(true, tlsCACerts)
	if err != nil {
		return nil, fmt.Errorf("failed to get tls cert pool: %w", err)
	}

	tlsServeCertPath := cmdutils.GetUserSetOptionalVarFromString(cmd, tlsServeCertPathFlagName, tlsServeCertPathEnvKey)

	tlsServeKeyPath := cmdutils.GetUserSetOptionalVarFromString(cmd, tlsServeKeyPathFlagName, tlsServeKeyPathFlagEnvKey)

	return &tlsParameters{
		systemCertPool: tlsSystemCertPool,
		serveCertPath:  tlsServeCertPath,
		serveKeyPath:   tlsServeKeyPath,
		tlsConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    rootCAs,
		},
	}, nil
}

func getRequestTokens(cmd *cobra.Command) map[string]string {
	requestTokens := cmdutils.GetUserSetOptionalVarFromArrayString(cmd, requestTokensFlagName,
		requestTokensEnvKey)

	tokens := make(map[string]string)

	for _, token := range requestTokens {
		split := strings.Split(token, "=")
		switch len(split) {
		case splitRequestTokenLength:
			tokens[split[0]] = split[1]
		default:
			logger.Warnf("invalid token '%s'", token)
		}
	}

	return tokens
}

func startService(params *serviceParameters, srv server) error { // nolint:funlen
	router := mux.NewRouter()

	provider, err := common.InitStore(params.dbParams, logger)
	if err != nil {
		return fmt.Errorf("failed to init provider: %w", err)
	}

	ariesConfig, err := newAriesConfig(params)
	if err != nil {
		return fmt.Errorf("failed to init aries config: %w", err)
	}

	// add health check endpoint
	healthCheckService := healthcheck.New()

	healthCheckHandlers := healthCheckService.GetOperations()
	for _, handler := range healthCheckHandlers {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	baseURL := params.baseURL
	if baseURL == "" {
		baseURL = params.host
	}

	ldStore, err := ld.NewStoreProvider(provider)
	if err != nil {
		return err
	}

	loader, err := ld.NewDocumentLoader(ldStore)
	if err != nil {
		return err
	}

	service, err := csh.New(&operation.Config{
		StoreProvider: provider,
		Aries:         ariesConfig,
		EDVClient:     adaptedEDVClientConstructor(),
		HTTPClient: &http.Client{Transport: &http.Transport{
			TLSClientConfig: params.tlsParams.tlsConfig,
		}},
		BaseURL:        baseURL,
		DIDDomain:      params.trustblocDomain,
		DocumentLoader: loader,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize confidential storage hub operations: %w", err)
	}

	for _, handler := range service.GetOperations() {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	for _, handler := range ldrest.New(ldsvc.New(ldStore)).GetRESTHandlers() {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	logger.Infof("starting server on host: %s", params.host)

	// start server on given port and serve using given handlers
	return srv.ListenAndServe(
		params.host,
		params.tlsParams.serveCertPath,
		params.tlsParams.serveKeyPath,
		cors.New(cors.Options{
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
			},
			AllowedHeaders: []string{
				"Origin",
				"Accept",
				"Content-Type",
				"X-Requested-With",
				"Authorization",
			},
		},
		).Handler(router))
}

// TODO make KMS and crypto configurable: https://github.com/trustbloc/ace/issues/578
func newAriesConfig(params *serviceParameters) (*operation.AriesConfig, error) {
	store, err := common.InitStore(params.dbParams, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init aries store: %w", err)
	}

	k, err := localkms.New(
		"local-lock://custom/primary/key/",
		&kmsProvider{
			sp: store,
			sl: &noop.NoLock{},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init local kms: %w", err)
	}

	c, err := tinkcrypto.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init tink crypto: %w", err)
	}

	didVDR, err := orb.New(
		nil,
		orb.WithDomain(params.trustblocDomain),
		orb.WithTLSConfig(params.tlsParams.tlsConfig),
		orb.WithAuthToken(params.requestTokens["sidetreeToken"]),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init trustbloc VDR: %w", err)
	}

	// TODO make these configurable:
	//  - DID resolvers
	//  - Key types
	//  - Verification method type
	return &operation.AriesConfig{
		KMS:    k,
		Crypto: c,
		WebKMS: func(url string, client webkms.HTTPClient, opts ...webkms.Opt) kms.KeyManager {
			return webkms.New(url, client, opts...)
		},
		WebCrypto: func(url string, client webcrypto.HTTPClient, opts ...webkms.Opt) crypto.Crypto {
			return webcrypto.New(url, client, opts...)
		},
		DIDResolvers: []zcapld2.DIDResolver{key.New(), didVDR},
		PublicDIDCreator: did.PublicDID(&did.Config{
			Method:                 params.identityDIDMethod,
			VerificationMethodType: "JsonWebKey2020",
			VDR:                    vdr.New(vdr.WithVDR(key.New()), vdr.WithVDR(didVDR)),
			JWKKeyCreator:          crypto2.JWKKeyCreator(kms.ED25519Type),
			CryptoKeyCreator:       crypto2.CryptoKeyCreator(kms.ED25519Type),
			DIDAnchorOrigin:        params.didAnchorOrigin,
		}),
	}, nil
}

type kmsProvider struct {
	sp ariesstorage.Provider
	sl secretlock.Service
}

func (k *kmsProvider) StorageProvider() ariesstorage.Provider {
	return k.sp
}

func (k *kmsProvider) SecretLock() secretlock.Service {
	return k.sl
}

func adaptedEDVClientConstructor() func(string, ...edv.Option) vault.ConfidentialStorageDocReader {
	return func(url string, opts ...edv.Option) vault.ConfidentialStorageDocReader {
		return &adaptedEDVClient{wrapped: edv.New(url, opts...)}
	}
}

type adaptedEDVClient struct {
	wrapped *edv.Client
}

func (a *adaptedEDVClient) ReadDocument(
	vaultID, docID string, opts ...edv.ReqOption) (*models.EncryptedDocument, error) {
	return a.wrapped.ReadDocument(vaultID, docID, opts...)
}
