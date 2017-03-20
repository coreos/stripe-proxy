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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

const separator = "_"

func computeMac(key, message []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}

func Sign(p *Permission, stripeKey []byte) (string, error) {
	permissionBytes, err := p.MarshalBinary()
	if err != nil {
		return "", err
	}

	mac := computeMac(stripeKey, permissionBytes)

	permissionEncoded := base64.RawStdEncoding.EncodeToString(permissionBytes)
	macEncoded := base64.RawStdEncoding.EncodeToString(mac)
	permissionAndMac := []string{permissionEncoded, macEncoded}

	return strings.Join(permissionAndMac, separator), nil
}

func Verify(credentials string, stripeKey []byte) (*Permission, error) {
	permissionAndMac := strings.SplitN(credentials, separator, 2)

	if len(permissionAndMac) != 2 {
		return nil, errors.New("Invalid signed permissions")
	}

	permissionBytes, err := base64.RawStdEncoding.DecodeString(permissionAndMac[0])
	if err != nil {
		return nil, err
	}
	expectedMac, err := base64.RawStdEncoding.DecodeString(permissionAndMac[1])
	if err != nil {
		return nil, err
	}

	computed := computeMac(stripeKey, permissionBytes)
	if !hmac.Equal(computed, expectedMac) {
		return nil, errors.New("MAC signature was not verified")
	}

	p := Permission{}
	err = p.BinaryUnmarshaler(permissionBytes)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
