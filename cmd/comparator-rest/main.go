/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package comparator-rest Comparator REST API.
//
//
// Terms Of Service:
//
//
//     Schemes: http, https
//     Version: 0.1.0
//     License: SPDX-License-Identifier: Apache-2.0
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package main

import (
	"github.com/spf13/cobra"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/ace/cmd/comparator-rest/startcmd"
)

var logger = log.New("comparator-rest")

func main() {
	rootCmd := &cobra.Command{
		Use: "comparator-rest",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	rootCmd.AddCommand(startcmd.GetStartCmd(&startcmd.HTTPServer{}))

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("execute root cmd: %s", err.Error())
	}
}
