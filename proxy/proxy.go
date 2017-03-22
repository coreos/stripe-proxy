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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go"
)

type ErrorResponse struct {
	StripeError stripe.Error `json:"error"`
}

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

func validButInsufficientError(msg string) *ErrorResponse {
	return &ErrorResponse{
		StripeError: stripe.Error{
			Type:           stripe.ErrorTypePermission,
			Msg:            msg,
			HTTPStatusCode: 403,
		}}
}

func invalidCredentialError(msg string) *ErrorResponse {
	return &ErrorResponse{
		StripeError: stripe.Error{
			Type:           stripe.ErrorTypeAuthentication,
			Msg:            msg,
			HTTPStatusCode: 403,
		}}
}

func checkPermissions(acc Access, res StripeResource, key []byte, req *http.Request) *ErrorResponse {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return invalidCredentialError("Request requires Authorization header")

	}

	// Check for bearer token
	signedPermissions := strings.TrimPrefix(authHeader, "Bearer ")
	if signedPermissions == authHeader {
		// Try basic auth
		var ok bool
		signedPermissions, _, ok = req.BasicAuth()
		if !ok {
			return invalidCredentialError("Request requires valid Basic or Bearer auth header")
		}
	}

	granted, err := Verify(signedPermissions, key)
	if err != nil {
		return invalidCredentialError(err.Error())
	}

	if !granted.Can(acc, res) {
		return validButInsufficientError("Request requires permission that was not granted")
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
					rw.WriteHeader(403)
					json.NewEncoder(rw).Encode(err)
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
