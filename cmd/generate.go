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
	"net/http"

	"github.com/spf13/cobra"
	"github.com/gorilla/mux"
	log "github.com/Sirupsen/logrus"
)

var port string

// generateCmd represents the serve command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Run the client UI webserver.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		r := mux.NewRouter()

		// Serve static client UI
		r.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			http.ServeFile(rw, req, "./static/index.html")
		})

		log.Infof("generate called on %s", port)
		log.Fatal(http.ListenAndServe(port, r))

		return nil
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVar(&port, "listen", ":9090", "Interface and port on which to listen")
}
