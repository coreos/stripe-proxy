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
	"testing"

	"github.com/stretchr/testify/assert"
)

const keyString = "pk_live_thisisateststripekey"

func TestGoodCredential(t *testing.T) {
	assert := assert.New(t)

	p := &Permission{}
	p.SetAccess(Read, ResourceCustomers, ResourceCharges, ResourceDisputes, ResourceEvents)
	p.SetAccess(Write, ResourceEvents)

	assert.NotZero(p.encoded)

	key := []byte(keyString)
	signed, err := Sign(p, key)
	assert.Nil(err)
	assert.NotZero(signed)

	q, err := Verify(signed, key)
	assert.Nil(err)
	assert.Equal(q, p)

	assert.True(q.Can(Write, ResourceEvents))
}

func TestBadCredential(t *testing.T) {
	assert := assert.New(t)

	p := &Permission{}
	p.SetAccess(Read, ResourceCustomers, ResourceCharges, ResourceDisputes, ResourceEvents)
	p.SetAccess(Write, ResourceEvents)

	assert.NotZero(p.encoded)

	key := []byte(keyString)
	signed, err := Sign(p, key)
	assert.Nil(err)
	assert.NotZero(signed)

	badSignature := signed[0 : len(signed)-1]

	q, err := Verify(badSignature, key)
	assert.Nil(q)
	assert.NotNil(err)
}
