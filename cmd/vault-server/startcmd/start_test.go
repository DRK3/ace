/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package startcmd

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListenAndServe(t *testing.T) {
	var w HTTPServer
	err := w.ListenAndServe("wronghost", "", "", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address wronghost: missing port in address")
}

func TestStartCmdWithBlankArg(t *testing.T) {
	t.Run("test blank host url arg", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{"--" + hostURLFlagName, ""}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Equal(t, "host-url value is empty", err.Error())
	})
}

func TestStartCmdWithMissingArg(t *testing.T) {
	t.Run("test missing host url arg", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		err := startCmd.Execute()

		require.Error(t, err)
		require.Equal(t,
			"Neither host-url (command line flag) nor VAULT_HOST_URL (environment variable) have been set.",
			err.Error())
	})
}

func TestStartCmdWithBlankEnvVar(t *testing.T) {
	t.Run("test blank host env var", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		err := os.Setenv(hostURLEnvKey, "")
		require.NoError(t, err)

		err = startCmd.Execute()
		require.Error(t, err)
		require.Equal(t, "VAULT_HOST_URL value is empty", err.Error())
	})
}

func TestStartCmdValidArgs(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	args := []string{
		"--" + hostURLFlagName, "localhost:8080",
		"--" + remoteKMSURLFlagName, "localhost:8081",
		"--" + edvURLFlagName, "localhost:8082",
		"--" + datasourceNameFlagName, "mem://test",
		"--" + requestTokensFlagName, "token2=tk2=1",
	}
	startCmd.SetArgs(args)

	err := startCmd.Execute()
	require.NoError(t, err)
}

func TestStartCmdEmptyDomain(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	startCmd.SetArgs([]string{
		"--" + hostURLFlagName, "localhost:8080",
		"--" + remoteKMSURLFlagName, "localhost:8081",
		"--" + edvURLFlagName, "localhost:8082",
		"--" + datasourceNameFlagName, "mem://test",
		"--" + didDomainFlagName, "",
	})

	err := startCmd.Execute()
	require.EqualError(t, err, "did-domain value is empty")
}

func TestStartCmdEmptyDidMethod(t *testing.T) {
	startCmd := GetStartCmd(&mockServer{})

	startCmd.SetArgs([]string{
		"--" + hostURLFlagName, "localhost:8080",
		"--" + remoteKMSURLFlagName, "localhost:8081",
		"--" + edvURLFlagName, "localhost:8082",
		"--" + datasourceNameFlagName, "mem://test",
		"--" + didMethodFlagName, "",
	})

	err := startCmd.Execute()
	require.EqualError(t, err, "did-method value is empty")
}

func TestDSN(t *testing.T) {
	t.Run("Unsupported driver", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + remoteKMSURLFlagName, "localhost:8081",
			"--" + edvURLFlagName, "localhost:8082",
			"--" + datasourceNameFlagName, "mem1://test",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported storage driver: mem1")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + remoteKMSURLFlagName, "localhost:8081",
			"--" + edvURLFlagName, "localhost:8082",
			"--" + datasourceNameFlagName, "mem",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid dbURL mem")
	})

	t.Run("Bad timeout", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + remoteKMSURLFlagName, "localhost:8081",
			"--" + edvURLFlagName, "localhost:8082",
			"--" + datasourceNameFlagName, "mem://test",
			"--" + datasourceTimeoutFlagName, "w1",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse dsn timeout")
	})
}

func TestTLSInvalidArgs(t *testing.T) {
	t.Run("test wrong tls cert pool flag", func(t *testing.T) {
		startCmd := GetStartCmd(&mockServer{})

		args := []string{
			"--" + hostURLFlagName, "localhost:8080",
			"--" + remoteKMSURLFlagName, "localhost:8081",
			"--" + edvURLFlagName, "localhost:8082",
			"--" + tlsSystemCertPoolFlagName, "wrong",
		}
		startCmd.SetArgs(args)

		err := startCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid syntax")
	})
}

func TestInitStore(t *testing.T) {
	provider, err := initStore("mongodb://dsn", 1, "prefix")
	require.NoError(t, err)
	require.NotNil(t, provider)
}

type mockServer struct{}

func (s *mockServer) ListenAndServe(host, certPath, keyPath string, handler http.Handler) error {
	return nil
}
