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

package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/coreos/stripe-proxy/proxy"
)

var upstreamURI string
var listenAddr string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the reverse proxy server.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		url, err := url.Parse(upstreamURI)
		if err != nil {
			return err
		}
		if url.Host == "" {
			msg := fmt.Sprintf("Unable to parse hostname from uri: %s", upstreamURI)
			return errors.New(msg)
		}
		log.Debugf("parsed host: %s %s", url.Host, err)

		rp := httputil.NewSingleHostReverseProxy(url)
		proxy := proxy.NewStripePermissionsProxy(stripeKey, rp)
		log.Infof("serve called with Stripe key: %s on %s", stripeKey, listenAddr)
		http.ListenAndServe(listenAddr, proxy)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&upstreamURI, "uri", "https://api.stripe.com", "Upstream Stripe API URI to talk to.")
	serveCmd.Flags().StringVar(&listenAddr, "listen", ":9090", "Interface and port on which to listen")
}
