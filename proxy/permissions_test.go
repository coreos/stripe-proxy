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

// Entitlement
type e struct {
	access   Access
	resource StripeResource
}

func TestPermissions(t *testing.T) {
	var permTests = []struct {
		encoded uint64
		allowed []e
		denied  []e
	}{
		{1,
			[]e{e{Read, ResourceAll}, e{Read, ResourceCustomers}, e{Read, ResourceBalance}},
			[]e{e{Write, ResourceAll}, e{ReadWrite, ResourceAll}, e{Write, ResourceCharges}}},
		{3,
			[]e{e{Read, ResourceAll}, e{Read, ResourceCustomers}, e{Read, ResourceBalance}, e{Write, ResourceAll}, e{ReadWrite, ResourceAll}, e{Write, ResourceCharges}},
			[]e{}},
		{4,
			[]e{e{Read, ResourceBalance}},
			[]e{e{Read, ResourceAll}, e{Read, ResourceCustomers}, e{Write, ResourceBalance}, e{Write, ResourceAll}, e{ReadWrite, ResourceAll}, e{Write, ResourceCharges}}},
		{8,
			[]e{e{Write, ResourceBalance}},
			[]e{e{Read, ResourceAll}, e{Read, ResourceCustomers}, e{Read, ResourceBalance}, e{Write, ResourceAll}, e{ReadWrite, ResourceAll}, e{Write, ResourceCharges}}},
		{864691128455135232,
			[]e{e{ReadWrite, ResourceRadarRule}, e{Read, ResourceRadarRule}, e{Write, ResourceRadarRule}},
			[]e{e{Read, ResourceAll}, e{Read, ResourceCustomers}, e{Read, ResourceBalance}, e{Write, ResourceAll}, e{ReadWrite, ResourceAll}, e{Write, ResourceCharges}}},
	}

	assert := assert.New(t)

	for _, tt := range permTests {
		p := Permission{tt.encoded}
		for _, allow := range tt.allowed {
			assert.True(p.Can(allow.access, allow.resource), "%b permission should allow %d to %d", p.encoded, allow.access, allow.resource)
		}
		for _, deny := range tt.denied {
			assert.False(p.Can(deny.access, deny.resource), "%b permission should deny %d to %d", p.encoded, deny.access, deny.resource)
		}
	}

}

func TestGrants(t *testing.T) {
	var createTests = []struct {
		grants   []e
		expected uint64
	}{
		{[]e{e{Read, ResourceAll}},
			1},
		{[]e{e{ReadWrite, ResourceAll}},
			3},
		{[]e{e{Read, ResourceFileUploads}, e{Write, ResourceCharges}},
			4128},
	}

	assert := assert.New(t)
	for _, tt := range createTests {
		p := Permission{}
		for _, toGrant := range tt.grants {
			p.SetAccess(toGrant.access, toGrant.resource)
		}
		assert.Equal(tt.expected, p.encoded)

		// At the very least we should have what was granted
		for _, toGrant := range tt.grants {
			assert.True(p.Can(toGrant.access, toGrant.resource))
		}
	}
}
