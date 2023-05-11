package feed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	providerMocks "github.com/chronicleprotocol/oracle-suite/pkg/price/provider/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name string
		path string
		test func(*testing.T, *Config)
	}{
		{
			name: "valid",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "key", cfg.EthereumKey)
				assert.Equal(t, uint32(60), cfg.Interval)
				expectedPairs := []provider.Pair{
					{Base: "ETH", Quote: "USD"},
					{Base: "BTC", Quote: "USD"},
				}
				assert.Equal(t, expectedPairs, cfg.Pairs)
			},
		},
		{
			name: "service",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				transport := local.New([]byte("test"), 1, nil)
				logger := null.New()
				keyRegistry := ethereum.KeyRegistry{
					"key": &ethereumMocks.Key{},
				}
				eventProvider := &providerMocks.Provider{}
				feed, err := cfg.Feed(Dependencies{
					KeysRegistry:  keyRegistry,
					PriceProvider: eventProvider,
					Transport:     transport,
					Logger:        logger,
				})
				require.NoError(t, err)
				assert.NotNil(t, feed)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cfg Config
			err := config.LoadFiles(&cfg, []string{"./testdata/" + test.path})
			require.NoError(t, err)
			test.test(t, &cfg)
		})
	}
}
