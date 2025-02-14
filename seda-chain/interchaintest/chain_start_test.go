package interchaintest

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/sedaprotocol/seda-chain/interchaintest/conformance"
)

func TestChainStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	numOfValidators := 2
	numOfFullNodes := 0

	chains := CreateChains(t, numOfValidators, numOfFullNodes)
	ic, ctx, _, _ := BuildAllChains(t, chains)

	chain := chains[0].(*cosmos.CosmosChain)

	userFunds := math.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	conformance.CosmWasm(t, ctx, chain, chainUser)

	require.NotNil(t, ic)
	require.NotNil(t, ctx)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
