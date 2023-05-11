//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package teleportevm

import (
	"errors"

	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const SignatureKey = "ethereum"

// Signer signs events using Ethereum signature.
//
// Signer could only sign events that have a "hash" field in the data. The
// value of that field is used to calculate the signature. The rest of the
// fields in the data are ignored. The calculated signature is stored in the
// "ethereum" field of the event's signatures map.
type Signer struct {
	signer wallet.Key
	types  []string
}

// NewSigner returns a new instance of the Signer struct.
func NewSigner(signer wallet.Key, types []string) *Signer {
	return &Signer{signer: signer, types: types}
}

// Sign implements the publisher.EventSigner interface.
func (l *Signer) Sign(event *messages.Event) (bool, error) {
	supports := false
	for _, t := range l.types {
		if t == event.Type {
			supports = true
			break
		}
	}
	if !supports {
		return false, nil
	}
	if event.Data == nil {
		return false, errors.New("event data is nil")
	}
	h, ok := event.Data["hash"]
	if !ok {
		return false, errors.New("missing hash field")
	}
	s, err := l.signer.SignMessage(h)
	if err != nil {
		return false, err
	}
	if event.Signatures == nil {
		event.Signatures = map[string]messages.EventSignature{}
	}
	event.Signatures[SignatureKey] = messages.EventSignature{
		Signer:    l.signer.Address().Bytes(),
		Signature: s.Bytes(),
	}
	return true, nil
}
