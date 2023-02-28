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

package origins

import (
	"encoding/json"
	"fmt"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
	"math/big"
	"time"

	"strings"
)

// GSU URL
const gsuURL = "https://api.gsucoin.app/Products/GSULive/?symbol=%s"
const wad = 1000000000000000000

type gsuResponse struct {
	Price     string    `json:"price"`
	Ask       string    `json:"ask"`
	Bid       string    `json:"bid"`
	Volume    string    `json:"volume"`
	Symbol    string    `json:"symbol"`
}

// GSU exchange handler
type GSU struct {
	WorkerPool query.WorkerPool
}

func (h *GSU) localPairName(pair Pair) string {
	return strings.ToUpper(pair.Base)
}

func (h *GSU) getURL(pairs []Pair) string {
	pairsStr := make([]string, len(pairs))
	for i, pair := range pairs {
		pairsStr[i] = h.localPairName(pair)
	}
	return fmt.Sprintf(gsuURL, strings.Join(pairsStr, ","))
}

func (h GSU) Pool() query.WorkerPool {
	return h.WorkerPool
}

func (h GSU) PullPrices(pairs []Pair) []FetchResult {
	crs, err := h.fetch(pairs)
	if err != nil {
		return fetchResultListWithErrors(pairs, err)
	}
	return crs
}

func (h *GSU) fetch(pairs []Pair) ([]FetchResult, error) {

	crs := make([]FetchResult, len(pairs))
	respMap := map[string]gsuResponse{}
	for i, pair := range pairs {
		symbol := h.localPairName(pair)
		url := fmt.Sprintf(gsuURL, symbol)
		req := &query.HTTPRequest{
			URL: url,
		}
		// make query
		res := h.Pool().Query(req)
		if res == nil {
			return nil, ErrEmptyOriginResponse
		}
		if res.Error != nil {
			return nil, res.Error
		}
		// parsing JSON
		var resp gsuResponse
		err := json.Unmarshal(res.Body, &resp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gsu response: %w", err)
		}
		respMap[symbol] = resp

		if resp, has := respMap[pair.Base]; has {
			p, err := h.newPrice(pair, resp)
			if err != nil {
				crs[i] = fetchResultWithError(
					pair,
					fmt.Errorf("failed to create price point from gsu response: %w: %s", err, res.Body),
				)
			} else {
				crs[i] = fetchResult(p)
			}
		} else {
			crs[i] = fetchResultWithError(
				pair,
				fmt.Errorf("failed to find symbol %s in gsu response: %s", pair, res.Body),
			)
		}
	}
	return crs, nil
}

func (h *GSU) newPrice(pair Pair, resp gsuResponse) (Price, error) {
	e := new(big.Float).SetUint64(wad)
	z := new(big.Int).SetUint64(0)

	// Parsing price from string.
	v, ok := new(big.Int).SetString(resp.price,10)
	if !ok {
		return Price{}, fmt.Errorf("failed to parse price from gsu exchange")
	}
	f := new(big.Float).SetInt(v)
	price, _ := f.Float64()

	//ask
	v, ok = new(big.Int).SetString(resp.ask,10)
	ask := float64(0)
	if ok && v.Cmp(z) != 0 {
		f = new(big.Float).SetInt(v)
		ask, _ = f.Float64()
	}

	//bid
	v, ok = new(big.Int).SetString(resp.bid,10)
	bid := float64(0)
	if ok && v.Cmp(z) != 0 {
		f = new(big.Float).SetInt(v)
		bid, _ = f.Float64()
	}

	//vol
	v, ok = new(big.Int).SetString(resp.volume,10)
	vol := float64(0)
	if ok && v.Cmp(z) != 0 {
		f = new(big.Float).SetInt(v)
		vol, _ = f.Float64()
	}

	// Building Price.
	return Price{
		Pair:      pair,
		Price:     price,
		Ask:       ask,
		Bid:       bid,
		Volume24h: vol,
		Timestamp: time.Now(),
	}, nil
}
