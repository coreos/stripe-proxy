// Copyright © 2017 stripe-proxy authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/coreos/stripe-proxy/proxy"
)

var inputToSign uint64

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign an integer which represents credentials to grant",
	Long: `
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("sign called with Stripe key: %s and input %b", stripeKey, inputToSign)

		p := proxy.NewPermission(inputToSign)
		signed, err := proxy.Sign(p, []byte(stripeKey))
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		log.Infof("Credentials:")
		fmt.Printf("%s\n", signed)
		log.Infof("Please copy and past the above credentials to your Stripe client")
	},
}

func init() {
	RootCmd.AddCommand(signCmd)
	signCmd.Flags().Uint64Var(&inputToSign, "input", 1, "Integer representation of permissions vector")
}
