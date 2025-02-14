package app

import (
	"encoding/json"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/sedaprotocol/seda-chain/app/params"
)

// distrModule wraps the x/distribution module to overwrite some genesis
// parameters.
type distrModule struct {
	distribution.AppModuleBasic
}

// DefaultGenesis returns custom x/distribution default genesis state.
func (distrModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesis := distrtypes.DefaultGenesisState()
	genesis.Params.CommunityTax = math.LegacyZeroDec()

	return cdc.MustMarshalJSON(genesis)
}

// mintModule wraps the x/mint module to overwrite some genesis
// parameters.
type mintModule struct {
	mint.AppModuleBasic
}

// DefaultGenesis returns custom x/mint default genesis state.
func (mintModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesis := minttypes.DefaultGenesisState()
	genesis.Params.MintDenom = params.DefaultBondDenom

	return cdc.MustMarshalJSON(genesis)
}
