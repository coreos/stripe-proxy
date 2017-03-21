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

package proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

// These routes will match in order, so the ResourceAll route is a fallback and
// transfer reversals will match before transfers.
var resourceRoutes = []struct {
	sr    StripeResource
	route string
}{
	{ResourceBalance, "/v1/balance"},
	{ResourceCharges, "/v1/charges"},
	{ResourceCustomers, "/v1/customers"},
	{ResourceDisputes, "/v1/disputes"},
	{ResourceEvents, "/v1/events"},
	{ResourceFileUploads, "/v1/files"},
	{ResourceRefunds, "/v1/refunds"},
	{ResourceTokens, "/v1/tokens"},
	{ResourceTransferReversals, "/v1/transfers/{transfer_id}/reversals"},
	{ResourceTransfers, "/v1/transfers"},
	{ResourceAll, "/v1/"},
}

var accessMethods = map[Access][]string{
	Read:  []string{"GET", "HEAD"},
	Write: []string{"POST", "DELETE", "PUT", "PATCH"},
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func stripAuth(req *http.Request) {
	req.Header.Del("Authorization")
}

func checkPermissions(acc Access, res StripeResource, key []byte, req *http.Request) error {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("Request requires Authorization header")

	}

	// Check for bearer token
	signedPermissions := strings.TrimPrefix(authHeader, "Bearer ")
	if signedPermissions == authHeader {
		// Try basic auth
		var ok bool
		signedPermissions, _, ok = req.BasicAuth()
		if !ok {
			return errors.New("Request requires valid Basic or Bearer auth header")
		}
	}

	granted, err := Verify(signedPermissions, key)
	if err != nil {
		return err
	}

	if !granted.Can(acc, res) {
		return errors.New("Request requires permission that was not granted")
	}

	return nil
}

func NewStripePermissionsProxy(stripeKey string, delegate http.Handler) http.Handler {
	r := mux.NewRouter()

	stripeKeyAsBytes := []byte(stripeKey)

	for _, rr := range resourceRoutes {
		for access, methods := range accessMethods {
			resourceToCheck := rr.sr
			accessToCheck := access

			f := func(rw http.ResponseWriter, req *http.Request) {
				err := checkPermissions(accessToCheck, resourceToCheck, stripeKeyAsBytes, req)
				if err != nil {
					// Abort the request
					http.Error(rw, err.Error(), http.StatusForbidden)
					return
				}

				req.SetBasicAuth(stripeKey, "")
				delegate.ServeHTTP(rw, req)
			}

			r.PathPrefix(rr.route).HandlerFunc(f).Methods(methods...)
		}
	}

	return r
}

func NewStripeScopedProxy(stripeAPI *url.URL) *httputil.ReverseProxy {
	stripeAPIQuery := stripeAPI.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = stripeAPI.Scheme
		req.URL.Host = stripeAPI.Host
		req.URL.Path = singleJoiningSlash(stripeAPI.Path, req.URL.Path)
		if stripeAPIQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = stripeAPIQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = stripeAPIQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

	}
	return &httputil.ReverseProxy{Director: director}
}
