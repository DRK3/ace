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
	"github.com/hyperledger/aries-framework-go/pkg/common/log"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	vdrpkg "github.com/hyperledger/aries-framework-go/pkg/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/httpbinding"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	cmdutils "github.com/trustbloc/edge-core/pkg/utils/cmd"
	tlsutils "github.com/trustbloc/edge-core/pkg/utils/tls"

	"github.com/trustbloc/ace/cmd/common"
	vaultclient "github.com/trustbloc/ace/pkg/client/vault"
	"github.com/trustbloc/ace/pkg/restapi/gatekeeper"
	"github.com/trustbloc/ace/pkg/restapi/gatekeeper/operation"
	"github.com/trustbloc/ace/pkg/restapi/gatekeeper/operation/vcprovider"
	"github.com/trustbloc/ace/pkg/restapi/healthcheck"
)

const (
	hostURLFlagName      = "host-url"
	hostURLFlagShorthand = "u"
	hostURLFlagUsage     = "Host URL to run the gatekeeper instance on. Format: HostName:Port."
	hostURLEnvKey        = "GK_HOST_URL"

	tlsSystemCertPoolFlagName  = "tls-systemcertpool"
	tlsSystemCertPoolFlagUsage = "Use system certificate pool." +
		" Possible values [true] [false]. Defaults to false if not set." +
		" Alternatively, this can be set with the following environment variable: " + tlsSystemCertPoolEnvKey
	tlsSystemCertPoolEnvKey = "GK_TLS_SYSTEMCERTPOOL"

	tlsCACertsFlagName  = "tls-cacerts"
	tlsCACertsFlagUsage = "Comma-Separated list of ca certs path." +
		" Alternatively, this can be set with the following environment variable: " + tlsCACertsEnvKey
	tlsCACertsEnvKey = "GK_TLS_CACERTS"

	tlsServeCertPathFlagName  = "tls-serve-cert"
	tlsServeCertPathFlagUsage = "Path to the server certificate to use when serving HTTPS." +
		" Alternatively, this can be set with the following environment variable: " + tlsServeCertPathEnvKey
	tlsServeCertPathEnvKey = "GK_TLS_SERVE_CERT"

	tlsServeKeyPathFlagName  = "tls-serve-key"
	tlsServeKeyPathFlagUsage = "Path to the private key to use when serving HTTPS." +
		" Alternatively, this can be set with the following environment variable: " + tlsServeKeyPathFlagEnvKey
	tlsServeKeyPathFlagEnvKey = "GK_TLS_SERVE_KEY"

	// did resolver url.
	didResolverURLFlagName  = "did-resolver-url"
	didResolverURLFlagUsage = "DID Resolver URL."
	didResolverURLEnvKey    = "GK_DID_RESOLVER_URL"

	blocDomainFlagName  = "bloc-domain"
	blocDomainFlagUsage = "Bloc domain"
	blocDomainEnvKey    = "GK_BLOC_DOMAIN"

	// remote JSON-LD context provider url.
	contextProviderFlagName  = "context-provider-url"
	contextProviderEnvKey    = "GK_CONTEXT_PROVIDER_URL"
	contextProviderFlagUsage = "Remote context provider URL to get JSON-LD contexts from." +
		" This flag can be repeated, allowing setting up multiple context providers." +
		" Alternatively, this can be set with the following environment variable (in CSV format): " +
		contextProviderEnvKey

	// vault server url.
	vaultServerURLFlagName  = "vault-server-url"
	vaultServerURLFlagUsage = "URL of the vault server. This field is mandatory."
	vaultServerURLEnvKey    = "GK_VAULT_SERVER_URL"

	// vc issuer server url.
	vcIssuerURLFlagName  = "vc-issuer-url"
	vcIssuerURLFlagUsage = "URL of the VC Issuer service. This field is mandatory."
	vcIssuerURLEnvKey    = "GK_VC_ISSUER_URL"

	vcRequestTokensFlagName  = "vc-request-tokens"
	vcRequestTokensEnvKey    = "GK_VC_REQUEST_TOKENS" //nolint:gosec
	vcRequestTokensFlagUsage = "Tokens used for http request to vc issuer" +
		" Alternatively, this can be set with the following environment variable: " + vcRequestTokensEnvKey

	tokenLength2 = 2
)

var logger = log.New("gatekeeper-rest")

type tlsParameters struct {
	systemCertPool bool
	caCerts        []string
	serveCertPath  string
	serveKeyPath   string
}

type serviceParameters struct {
	host                string
	tlsParams           *tlsParameters
	dbParams            *common.DBParameters
	blocDomain          string
	didResolverURL      string
	contextProviderURLs []string
	vcRequestTokens     map[string]string
	vcIssuerURL         string
	vaultServerURL      string
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

	tlsServeCertPath := cmdutils.GetUserSetOptionalVarFromString(cmd, tlsServeCertPathFlagName, tlsServeCertPathEnvKey)

	tlsServeKeyPath := cmdutils.GetUserSetOptionalVarFromString(cmd, tlsServeKeyPathFlagName, tlsServeKeyPathFlagEnvKey)

	return &tlsParameters{
		systemCertPool: tlsSystemCertPool,
		caCerts:        tlsCACerts,
		serveCertPath:  tlsServeCertPath,
		serveKeyPath:   tlsServeKeyPath,
	}, nil
}

func createStartCmd(srv server) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts Gatekeeper server",
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

	tlsParams, err := getTLS(cmd)
	if err != nil {
		return nil, err
	}

	dbParams, err := common.DBParams(cmd)
	if err != nil {
		return nil, err
	}

	blocDomain, err := cmdutils.GetUserSetVarFromString(cmd, blocDomainFlagName, blocDomainEnvKey, true)
	if err != nil {
		return nil, err
	}

	didResolverURL, err := cmdutils.GetUserSetVarFromString(cmd,
		didResolverURLFlagName, didResolverURLEnvKey, true)
	if err != nil {
		return nil, err
	}

	contextProviderURLs, err := cmdutils.GetUserSetVarFromArrayString(cmd, contextProviderFlagName,
		contextProviderEnvKey, true)
	if err != nil {
		return nil, err
	}

	vaultServerURL, err := cmdutils.GetUserSetVarFromString(cmd, vaultServerURLFlagName,
		vaultServerURLEnvKey, false)
	if err != nil {
		return nil, err
	}

	vcIssuerURL, err := cmdutils.GetUserSetVarFromString(cmd, vcIssuerURLFlagName, vcIssuerURLEnvKey, false)
	if err != nil {
		return nil, err
	}

	vcRequestTokens, err := getVCRequestTokens(cmd)
	if err != nil {
		return nil, err
	}

	return &serviceParameters{
		host:                host,
		tlsParams:           tlsParams,
		dbParams:            dbParams,
		blocDomain:          blocDomain,
		didResolverURL:      didResolverURL,
		contextProviderURLs: contextProviderURLs,
		vcRequestTokens:     vcRequestTokens,
		vcIssuerURL:         vcIssuerURL,
		vaultServerURL:      vaultServerURL,
	}, err
}

func createFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(hostURLFlagName, hostURLFlagShorthand, "", hostURLFlagUsage)
	cmd.Flags().StringP(tlsSystemCertPoolFlagName, "", "", tlsSystemCertPoolFlagUsage)
	cmd.Flags().StringArrayP(tlsCACertsFlagName, "", []string{}, tlsCACertsFlagUsage)
	cmd.Flags().StringP(tlsServeCertPathFlagName, "", "", tlsServeCertPathFlagUsage)
	cmd.Flags().StringP(tlsServeKeyPathFlagName, "", "", tlsServeKeyPathFlagUsage)
	cmd.Flags().StringP(blocDomainFlagName, "", "", blocDomainFlagUsage)
	cmd.Flags().StringP(didResolverURLFlagName, "", "", didResolverURLFlagUsage)
	cmd.Flags().StringArrayP(contextProviderFlagName, "", []string{}, contextProviderFlagUsage)
	cmd.Flags().StringP(vaultServerURLFlagName, "", "", vaultServerURLFlagUsage)
	cmd.Flags().StringP(vcIssuerURLFlagName, "", "", vcIssuerURLFlagUsage)
	cmd.Flags().StringArrayP(vcRequestTokensFlagName, "", []string{}, vcRequestTokensFlagUsage)

	common.Flags(cmd)
}

func startService(params *serviceParameters, srv server) error { // nolint: funlen
	rootCAs, err := tlsutils.GetCertPool(params.tlsParams.systemCertPool, params.tlsParams.caCerts)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{RootCAs: rootCAs, MinVersion: tls.VersionTLS12}

	storeProvider, err := common.InitStore(params.dbParams, logger)
	if err != nil {
		return err
	}

	router := mux.NewRouter()

	// add health check endpoint
	healthCheckService := healthcheck.New()

	healthCheckHandlers := healthCheckService.GetOperations()
	for _, handler := range healthCheckHandlers {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: tlsConfig,
	}}

	vdri, err := createVDRI(params.didResolverURL, params.blocDomain, httpClient)
	if err != nil {
		return err
	}

	ldStore, err := common.CreateLDStoreProvider(storeProvider)
	if err != nil {
		return err
	}

	documentLoader, err := common.CreateJSONLDDocumentLoader(ldStore, httpClient, params.contextProviderURLs)
	if err != nil {
		return err
	}

	vClient := vaultclient.New(params.vaultServerURL, vaultclient.WithHTTPClient(httpClient))

	vcProvider := vcprovider.New(&vcprovider.Config{
		VCIssuerURL:     params.vcIssuerURL,
		VCRequestTokens: params.vcRequestTokens,
		DocumentLoader:  documentLoader,
		HTTPClient:      httpClient,
	})

	service, err := gatekeeper.New(&operation.Config{
		StorageProvider: storeProvider,
		VaultClient:     vClient,
		VDRI:            vdri,
		VCProvider:      vcProvider,
	})
	if err != nil {
		return err
	}

	for _, handler := range service.GetOperations() {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	// start server on given port and serve using given handlers
	return srv.ListenAndServe(params.host, params.tlsParams.serveCertPath, params.tlsParams.serveKeyPath,
		cors.New(cors.Options{
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodDelete,
			},
			AllowedHeaders: []string{
				"Origin",
				"Accept",
				"Content-Type",
				"X-Requested-With",
				"Authorization",
			},
		}).Handler(router))
}

func getVCRequestTokens(cmd *cobra.Command) (map[string]string, error) {
	requestTokens, err := cmdutils.GetUserSetVarFromArrayString(cmd, vcRequestTokensFlagName,
		vcRequestTokensEnvKey, true)
	if err != nil {
		return nil, err
	}

	tokens := make(map[string]string)

	for _, token := range requestTokens {
		split := strings.Split(token, "=")
		switch len(split) {
		case tokenLength2:
			tokens[split[0]] = split[1]
		default:
			logger.Warnf("invalid token '%s'", token)
		}
	}

	return tokens, nil
}

func createVDRI(didResolverURL, blocDomain string, httpClient *http.Client) (vdrapi.Registry, error) { //nolint:ireturn
	var opts []vdrpkg.Option

	if didResolverURL != "" {
		didResolverVDRI, err := httpbinding.New(didResolverURL, httpbinding.WithHTTPClient(httpClient),
			httpbinding.WithAccept(func(method string) bool {
				return method == "orb" || method == "v1" || method == "elem" || method == "sov" ||
					method == "web" || method == "key" || method == "factom"
			}))
		if err != nil {
			return nil, fmt.Errorf("failed to create new universal resolver vdr: %w", err)
		}

		// add universal resolver vdr
		opts = append(opts, vdrpkg.WithVDR(didResolverVDRI))
	}

	if blocDomain != "" {
		vdr, err := orb.New(nil, orb.WithDomain(blocDomain), orb.WithHTTPClient(httpClient))
		if err != nil {
			return nil, err
		}

		opts = append(opts, vdrpkg.WithVDR(vdr))
	}

	return vdrpkg.New(opts...), nil
}
