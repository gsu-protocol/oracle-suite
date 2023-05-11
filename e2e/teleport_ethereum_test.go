package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Teleport_Ethereum(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	mocks := []*smocker.Mock{
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_blockNumber")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_blockNumber.json")).
			Mock(),
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_getBlockByNumber")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_getBlockByNumber.json")).
			Mock(),
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_getLogs")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_getLogs.json")).
			Mock(),
	}

	require.NoError(t, s.AddMocks(ctx, mocks))

	cmd1 := command(ctx, "..", nil, "./lair", "run", "-c", "./e2e/testdata/config/lair.hcl", "-v", "debug")
	cmd2 := command(ctx, "..", nil, "./leeloo", "run", "-c", "./e2e/testdata/config/leeloo_ethereum.hcl", "-v", "debug")
	cmd3 := command(ctx, "..", nil, "./leeloo", "run", "-c", "./e2e/testdata/config/leeloo2_ethereum.hcl", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = cmd1.Wait()
		_ = cmd2.Wait()
		_ = cmd3.Wait()
	}()

	// Start the lair and wait for it to be ready.
	if err := cmd1.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30100)

	// Start the leeloo nodes and wait for them to be ready.
	// Signing leeloo events requires a lot of memory, if two instances are started at the same time
	// it may happen, that both instances will try to sign the same event at the same time which
	// may cause a OOM error on a staging environment. Because of that, we start the second instance
	// with a 5-second delay.
	if err := cmd2.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	time.Sleep(5 * time.Second)
	if err := cmd3.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30101)
	waitForPort(ctx, "localhost", 30102)

	lairResponse, err := waitForLair(ctx, "http://localhost:30000/?type=teleport_evm&index=0x5f4a7c89123ed655b7fce471f2f14a4b699a9edfabeef6a8d5571976907f1884", 2)
	if err != nil {
		require.Fail(t, err.Error())
	}

	require.Len(t, lairResponse, 2)

	assert.Equal(t,
		"52494e4b4542592d534c4156452d415242495452554d2d31000000000000000052494e4b4542592d4d41535445522d3100000000000000000000000000000000000000000000000000000000d747d98b8a2b28dfd6cd9f0e6015ad2a671611180000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000008180000000000000000000000000000000000000000000000000000000062b1e05f",
		lairResponse[0].Data["event"],
	)
	assert.Equal(t,
		"36223c6974790ab39f3b094fccbeb05b60592983206bddbb9c5fc9d9ede4706f",
		lairResponse[0].Data["hash"],
	)
	assert.Equal(t,
		"2d800d93b065ce011af83f316cef9f0d005b0aa4",
		lairResponse[0].Signatures["ethereum"].Signer,
	)
	assert.Equal(t,
		"e68c360c2c3eb0452369b8829611e2587896e1d990e3924cb6d18c178afda5735eeb99b424ba1f8230b3005f937705743885f8413c249514d8727514c3b324671c",
		lairResponse[0].Signatures["ethereum"].Signature,
	)

	assert.Equal(t,
		"52494e4b4542592d534c4156452d415242495452554d2d31000000000000000052494e4b4542592d4d41535445522d3100000000000000000000000000000000000000000000000000000000d747d98b8a2b28dfd6cd9f0e6015ad2a671611180000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000008180000000000000000000000000000000000000000000000000000000062b1e05f",
		lairResponse[1].Data["event"],
	)
	assert.Equal(t,
		"36223c6974790ab39f3b094fccbeb05b60592983206bddbb9c5fc9d9ede4706f",
		lairResponse[1].Data["hash"],
	)
	assert.Equal(t,
		"e3ced0f62f7eb2856d37bed128d2b195712d2644",
		lairResponse[1].Signatures["ethereum"].Signer,
	)
	assert.Equal(t,
		"36913257c92c309bcbf415a2a041ba1eeb02117c64e59aa73c54ddaee97126ec7b091cf83d65e912bd6d2dbb306a42e466a7080111cc797dd78b621df918b8aa1b",
		lairResponse[1].Signatures["ethereum"].Signature,
	)
}

func Test_Teleport_Ethereum_Replay(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	mocks := []*smocker.Mock{
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_blockNumber")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_blockNumber.json")).
			Mock(),
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_getBlockByNumber")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_getBlockByNumber.json")).
			Mock(),
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_getLogs")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_getLogs.json")).
			Mock(),
	}

	require.NoError(t, s.AddMocks(ctx, mocks))

	eventDate := time.Unix(1655824479, 0)
	replayAfter := time.Since(eventDate) + 40*time.Second

	cmd1 := command(ctx, "..", nil, "./lair", "run", "-c", "./e2e/testdata/config/lair_test_replay.hcl", "-v", "debug")
	cmd2 := command(ctx, "..", []string{fmt.Sprintf("REPLAY_AFTER=%d", int(replayAfter.Seconds()))}, "./leeloo", "run", "-c", "./e2e/testdata/config/leeloo_ethereum.hcl", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = cmd1.Wait()
		_ = cmd2.Wait()
	}()

	// Run Lair with 20-second delay, so the only way to receive event is to wait
	// for it to be replayed.
	go func() {
		time.Sleep(20 * time.Second)
		if err := cmd1.Start(); err != nil {
			require.Fail(t, err.Error())
		}
	}()
	if err := cmd2.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30100)
	waitForPort(ctx, "localhost", 30101)

	lairResponse, err := waitForLair(ctx, "http://localhost:30000/?type=teleport_evm&index=0x5f4a7c89123ed655b7fce471f2f14a4b699a9edfabeef6a8d5571976907f1884", 1)
	if err != nil {
		require.Fail(t, err.Error())
	}

	require.Len(t, lairResponse, 1)
}
