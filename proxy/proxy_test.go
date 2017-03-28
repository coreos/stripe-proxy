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
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
)

const proxyTestStripeKey = "pk_live_thisisateststripekey"

type TeapotUpstream struct {
	mock.Mock
}

func (m *TeapotUpstream) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.Called()

	err := ErrorResponse{
		StripeError: stripe.Error{
			Type:           stripe.ErrorTypeInvalidRequest,
			Msg:            "I'm a teapot",
			HTTPStatusCode: 418,
		}}

	rw.WriteHeader(418)
	json.NewEncoder(rw).Encode(err)
}

func getBackends(s *httptest.Server) *stripe.Backends {
	httpClient := &http.Client{Timeout: 1 * time.Second}
	apiUrl := s.URL + "/v1"
	return &stripe.Backends{
		API: stripe.BackendConfiguration{
			stripe.APIBackend, apiUrl, httpClient},
		Uploads: stripe.BackendConfiguration{
			stripe.UploadsBackend, apiUrl, httpClient},
	}
}

func newTeapotProxy() (http.Handler, *TeapotUpstream) {
	testUpstream := new(TeapotUpstream)
	testUpstream.On("ServeHTTP").Return()
	permProxy := NewStripePermissionsProxy(proxyTestStripeKey, testUpstream)
	return permProxy, testUpstream
}

func TestPassingRequest(t *testing.T) {
	assert := assert.New(t)

	proxy, testUpstream := newTeapotProxy()
	server := httptest.NewServer(proxy)
	defer server.Close()

	p := &Permission{}
	p.SetAccess(Read, ResourceCustomers)
	signed, err := Sign(p, []byte(proxyTestStripeKey))
	assert.Nil(err)

	sc := &client.API{}
	sc.Init(signed, getBackends(server))
	custlist := sc.Customers.List(nil)

	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 1)

	stripeError, ok := custlist.Err().(*stripe.Error)
	assert.True(ok)
	assert.Equal(418, stripeError.HTTPStatusCode)
}

func TestRejectedBadCredential(t *testing.T) {
	assert := assert.New(t)

	proxy, testUpstream := newTeapotProxy()
	server := httptest.NewServer(proxy)
	defer server.Close()

	p := &Permission{}
	p.SetAccess(Read, ResourceCustomers)
	signed, err := Sign(p, []byte(proxyTestStripeKey))
	assert.Nil(err)

	sc := &client.API{}
	sc.Init(signed[0:len(signed)-1], getBackends(server))
	custlist := sc.Customers.List(nil)

	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 0)

	stripeError, ok := custlist.Err().(*stripe.Error)
	assert.True(ok)

	assert.Equal(403, stripeError.HTTPStatusCode)
	assert.Equal(stripe.ErrorTypeAuthentication, stripeError.Type)
}

func TestRejectedPermissions(t *testing.T) {
	assert := assert.New(t)

	proxy, testUpstream := newTeapotProxy()
	server := httptest.NewServer(proxy)
	defer server.Close()

	p := &Permission{}
	p.SetAccess(Read, ResourceCustomers)
	signed, err := Sign(p, []byte(proxyTestStripeKey))
	assert.Nil(err)

	sc := &client.API{}
	sc.Init(signed, getBackends(server))
	_, deleteErr := sc.Customers.Del("cus_fakecustid")

	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 0)

	stripeDelError, ok := deleteErr.(*stripe.Error)
	assert.True(ok)

	assert.Equal(403, stripeDelError.HTTPStatusCode)
	assert.Equal(stripe.ErrorTypePermission, stripeDelError.Type)

	xfrListErr := sc.Transfers.List(nil)
	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 0)

	stripeXfrError, ok := xfrListErr.Err().(*stripe.Error)
	assert.True(ok)

	assert.Equal(403, stripeXfrError.HTTPStatusCode)
	assert.Equal(stripe.ErrorTypePermission, stripeXfrError.Type)

	// Grant read on transfers now
	p.SetAccess(Read, ResourceTransfers)
	newGrant, err := Sign(p, []byte(proxyTestStripeKey))
	assert.Nil(err)

	sc.Init(newGrant, getBackends(server))

	xfrList := sc.Transfers.List(nil)
	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 1)

	stripeError, ok := xfrList.Err().(*stripe.Error)
	assert.True(ok)
	assert.Equal(418, stripeError.HTTPStatusCode)
}

func TestExpandParams(t *testing.T) {
	assert := assert.New(t)

	proxy, testUpstream := newTeapotProxy()
	server := httptest.NewServer(proxy)
	defer server.Close()

	p := &Permission{}
	p.SetAccess(Read, ResourceCharges)
	signed, err := Sign(p, []byte(proxyTestStripeKey))
	assert.Nil(err)

	sc := &client.API{}
	sc.Init(signed, getBackends(server))

	// Without the expand it works
	params := &stripe.ChargeParams{}
	_, ch := sc.Charges.Get("ch_example_id", params)

	teapotError, ok := ch.(*stripe.Error)
	assert.True(ok)
	assert.Equal(418, teapotError.HTTPStatusCode)
	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 1)

	// With the expand it fails
	params.Expand("customer")
	_, expectError := sc.Charges.Get("ch_example_id", params)

	permissionError, ok := expectError.(*stripe.Error)
	assert.True(ok)

	assert.Equal(403, permissionError.HTTPStatusCode)
	assert.Equal(stripe.ErrorTypePermission, permissionError.Type)
	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 1)

	// By granting resource all it works again
	p.SetAccess(Read, ResourceAll)
	newaccess, err := Sign(p, []byte(proxyTestStripeKey))
	sc.Init(newaccess, getBackends(server))

	_, chWorks := sc.Charges.Get("ch_example_id", params)
	expectTeapotAgain, ok := chWorks.(*stripe.Error)
	assert.True(ok)
	assert.Equal(418, expectTeapotAgain.HTTPStatusCode)
	testUpstream.AssertNumberOfCalls(t, "ServeHTTP", 2)
}
