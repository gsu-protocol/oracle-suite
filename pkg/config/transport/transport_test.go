package transport

import (
	"testing"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
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
				assert.NotNil(t, cfg.LibP2P)
				assert.NotNil(t, cfg.WebAPI)

				// LibP2P
				assert.Equal(t, "0x1234567890123456789012345678901234567890", cfg.LibP2P.Feeds[0].String())
				assert.Equal(t, "0x2345678901234567890123456789012345678901", cfg.LibP2P.Feeds[1].String())
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/6000"}, cfg.LibP2P.ListenAddrs)
				assert.Equal(t, "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff", cfg.LibP2P.PrivKeySeed)
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/7000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi"}, cfg.LibP2P.BootstrapAddrs)
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi"}, cfg.LibP2P.DirectPeersAddrs)
				assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/9000"}, cfg.LibP2P.BlockedAddrs)
				assert.Equal(t, true, cfg.LibP2P.DisableDiscovery)
				assert.Equal(t, "key", cfg.LibP2P.EthereumKey)

				// WebAPI
				assert.Equal(t, "0x3456789012345678901234567890123456789012", cfg.WebAPI.Feeds[0].String())
				assert.Equal(t, "0x4567890123456789012345678901234567890123", cfg.WebAPI.Feeds[1].String())
				assert.Equal(t, "localhost:8080", cfg.WebAPI.ListenAddr)
				assert.Equal(t, "localhost:9050", cfg.WebAPI.Socks5ProxyAddr)
				assert.Equal(t, "key", cfg.WebAPI.EthereumKey)
				assert.NotNil(t, cfg.WebAPI.EthereumAddressBook)
				assert.NotNil(t, cfg.WebAPI.StaticAddressBook)

				// EthereumAddressBook
				assert.Equal(t, "0x5678901234567890123456789012345678901234", cfg.WebAPI.EthereumAddressBook.ContractAddr.String())
				assert.Equal(t, "client", cfg.WebAPI.EthereumAddressBook.EthereumClient)

				// StaticAddressBook
				assert.Equal(t, []string{"https://example.com/api/v1/endpoint"}, cfg.WebAPI.StaticAddressBook.Addresses)
			},
		},
		{
			name: "service",
			path: "config.hcl",
			test: func(t *testing.T, cfg *Config) {
				key := &mocks.Key{}
				key.On("Address").Return(types.AddressFromHex("0x1234567890123456789012345678901234567890"))
				keyRegistry := ethereum.KeyRegistry{
					"key": key,
				}
				clientRegistry := ethereum.ClientRegistry{
					"client": &mocks.RPC{},
				}
				transport, err := cfg.Transport(Dependencies{
					Keys:     keyRegistry,
					Clients:  clientRegistry,
					Messages: nil,
					Logger:   null.New(),
				})
				require.NoError(t, err)
				assert.NotNil(t, transport)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			var cfg Config
			err := config.LoadFiles(&cfg, []string{"./testdata/" + test.path})
			require.NoError(t, err)
			test.test(t, &cfg)
		})
	}
}
