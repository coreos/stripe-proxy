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
	"encoding/binary"
)

type StripeResource int

// Note: these do not use iota so that they are stable through modifications
// of the list.
const (
	ResourceAll StripeResource = 0

	// Core resources
	ResourceBalance           = 1
	ResourceCharges           = 2
	ResourceCustomers         = 3
	ResourceDisputes          = 4
	ResourceEvents            = 5
	ResourceFileUploads       = 6
	ResourceRefunds           = 7
	ResourceTokens            = 8
	ResourceTransfers         = 9
	ResourceTransferReversals = 10

	// Connect resources
	ResourceAccount              = 11
	ResourceApplicationFeeRefund = 12
	ResourceApplicationFee       = 13
	ResourceRecipient            = 14
	ResourceCountrySpec          = 15
	ResourceExternalAccount      = 16

	// Payment methods
	ResourceSource = 17

	// Relay resources
	ResourceOrder       = 18
	ResourceOrderReturn = 19
	ResourceProduct     = 20
	ResourceSKU         = 21

	// Subscription resources
	ResourceCoupon           = 22
	ResourceInvoice          = 23
	ResourceInvoiceItem      = 24
	ResourcePlan             = 25
	ResourceSubscription     = 26
	ResourceSubscriptionItem = 27

	// Radar resources
	ResourceRadarReview = 28
	ResourceRadarRule   = 29
)

type Access int

// Note: these do not use iota so that they are stable through modifications
// of the list.
const (
	None      = 0
	Read      = 1
	Write     = 2
	ReadWrite = 3
)

type Permission struct {
	encoded uint64
}

func NewPermission(initialValue uint64) *Permission {
	return &Permission{initialValue}
}

func (p *Permission) MarshalBinary() ([]byte, error) {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, p.encoded)
	return bs, nil
}

func (p *Permission) BinaryUnmarshaler(data []byte) error {
	p.encoded = binary.BigEndian.Uint64(data)
	return nil
}

func resourceMask(access Access, resources ...StripeResource) uint64 {
	var mask uint64
	for _, resource := range resources {
		mask |= uint64(access << (uint64(resource) * 2))
	}
	return mask
}

func (p *Permission) Can(access Access, resources ...StripeResource) bool {
	mask := resourceMask(access, resources...)
	allMask := resourceMask(access, ResourceAll)
	return mask&p.encoded == mask || allMask&p.encoded == allMask
}

func (p *Permission) SetAccess(access Access, resources ...StripeResource) {
	p.encoded |= resourceMask(access, resources...)
}
