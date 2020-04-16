package exchange

import (
	"makerdao/gofer"
	"makerdao/gofer/query"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// mockWorkerPool mock worker pool implementation for tests
type mockWorkerPool struct {
	resp *query.HTTPResponse
}

func newMockWorkerPool(resp *query.HTTPResponse) *mockWorkerPool {
	return &mockWorkerPool{
		resp: resp,
	}
}

func (mwp *mockWorkerPool) Start() {}

func (mwp *mockWorkerPool) Stop() error {
	return nil
}

func (mwp *mockWorkerPool) Query(req *query.HTTPRequest) *query.HTTPResponse {
	return mwp.resp
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ExchangesSuite struct {
	suite.Suite
	pool query.WorkerPool
}

func (suite *ExchangesSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *ExchangesSuite) TestCallErrorNegative() {
	pool := newMockWorkerPool(nil)

	res, err := Call(nil, nil)
	assert.Nil(suite.T(), res)
	assert.Error(suite.T(), err)

	res, err = Call(pool, nil)
	assert.Nil(suite.T(), res)
	assert.Error(suite.T(), err)

	res, err = Call(pool, &gofer.PotentialPricePoint{})
	assert.Nil(suite.T(), res)
	assert.Error(suite.T(), err)

	pp := &gofer.PotentialPricePoint{
		Exchange: &gofer.Exchange{
			Name: "unknown",
		},
	}
	res, err = Call(pool, pp)
	assert.Nil(suite.T(), res)
	assert.Error(suite.T(), err)
}

func (suite *ExchangesSuite) TestFailWithNilResponseForBinance() {
	pool := newMockWorkerPool(nil)
	pp := &gofer.PotentialPricePoint{
		Exchange: &gofer.Exchange{
			Name: "binance",
		},
		Pair: &gofer.Pair{
			Base:  "BTC",
			Quote: "ETH",
		},
	}

	res, err := Call(pool, pp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), res)
}

func (suite *ExchangesSuite) TestSuccessBinance() {
	resp := &query.HTTPResponse{
		Body:  []byte(`{"symbol":"ETHBTC","price":"0.02436100"}`),
		Error: nil,
	}
	p := &gofer.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pool := newMockWorkerPool(resp)
	pp := &gofer.PotentialPricePoint{
		Exchange: &gofer.Exchange{
			Name: "binance",
		},
		Pair: p,
	}

	res, err := Call(pool, pp)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), p, res.Pair)
	assert.EqualValues(suite.T(), 0.024361, res.Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExchangesSuite(t *testing.T) {
	suite.Run(t, new(ExchangesSuite))
}
