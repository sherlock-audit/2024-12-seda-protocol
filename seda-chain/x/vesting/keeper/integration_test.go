package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkintegration "github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/integration"
	pubkeykeeper "github.com/sedaprotocol/seda-chain/x/pubkey/keeper"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	"github.com/sedaprotocol/seda-chain/x/staking"
	stakingkeeper "github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/vesting"
	"github.com/sedaprotocol/seda-chain/x/vesting/types"
)

const (
	bech32Prefix = "seda"
	bondDenom    = "aseda"
)

var (
	zeroCoins  sdk.Coins
	funderAddr = sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d")
	testAddrs  = []sdk.AccAddress{
		sdk.AccAddress([]byte("to0_________________")),
		sdk.AccAddress([]byte("to1_________________")),
		sdk.AccAddress([]byte("to2_________________")),
		sdk.AccAddress([]byte("to3_________________")),
		sdk.AccAddress([]byte("to4_________________")),
		sdk.AccAddress([]byte("to5_________________")),
		sdk.AccAddress([]byte("to6_________________")),
		sdk.AccAddress([]byte("to7_________________")),
		sdk.AccAddress([]byte("to8_________________")),
		sdk.AccAddress([]byte("to9_________________")),
	}
)

type fixture struct {
	*integration.IntegationApp
	cdc           codec.Codec
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, vesting.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := sdkintegration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{Time: time.Now().UTC()}, true, logger)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	maccPerms := map[string][]string{
		minttypes.ModuleName:              {authtypes.Minter},
		sdkstakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		sdkstakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		types.ModuleName:                  {authtypes.Burner},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(params.Bech32PrefixAccAddr),
		params.Bech32PrefixAccAddr,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	sdkStakingKeeper := sdkstakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[sdkstakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(params.Bech32PrefixValAddr), addresscodec.NewBech32Codec(params.Bech32PrefixConsAddr))
	stakingKeeper := stakingkeeper.NewKeeper(sdkStakingKeeper, addresscodec.NewBech32Codec(params.Bech32PrefixValAddr))

	stakingParams := sdkstakingtypes.DefaultParams()
	stakingParams.BondDenom = bondDenom
	err := stakingKeeper.SetParams(newCtx, stakingParams)
	require.NoError(tb, err)

	slashingKeeper := slashingkeeper.NewKeeper(
		cdc,
		nil,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	pubKeyKeeper := pubkeykeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[pubkeytypes.StoreKey]),
		stakingKeeper,
		slashingKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
		authtypes.NewModuleAddress("gov").String(),
	)
	stakingKeeper.SetPubKeyKeeper(pubKeyKeeper)

	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, pubKeyKeeper)
	vestingModule := vesting.NewAppModule(accountKeeper, bankKeeper, stakingKeeper)

	integrationApp := integration.NewIntegrationApp(
		newCtx, logger, keys, cdc,
		baseapp.NewMsgServiceRouter(),
		map[string]appmodule.AppModule{
			authtypes.ModuleName:       authModule,
			banktypes.ModuleName:       bankModule,
			sdkstakingtypes.ModuleName: stakingModule,
			types.ModuleName:           vestingModule,
		},
	)

	return &fixture{
		IntegationApp: integrationApp,
		cdc:           cdc,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: *stakingKeeper,
	}
}

func createValidators(t *testing.T, f *fixture, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress, []cryptotypes.PubKey) {
	t.Helper()
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.Context(), 5, math.NewInt(5e18))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	valPks := simtestutil.CreateTestPubKeys(5)

	val1, err := sdkstakingtypes.NewValidator(valAddrs[0].String(), valPks[0], sdkstakingtypes.Description{})
	require.NoError(t, err)
	val2, err := sdkstakingtypes.NewValidator(valAddrs[1].String(), valPks[1], sdkstakingtypes.Description{})
	require.NoError(t, err)
	val3, err := sdkstakingtypes.NewValidator(valAddrs[2].String(), valPks[2], sdkstakingtypes.Description{})
	require.NoError(t, err)

	require.NoError(t, f.stakingKeeper.SetValidator(f.Context(), val1))
	require.NoError(t, f.stakingKeeper.SetValidator(f.Context(), val2))
	require.NoError(t, f.stakingKeeper.SetValidator(f.Context(), val3))
	require.NoError(t, f.stakingKeeper.SetValidatorByConsAddr(f.Context(), val1))
	require.NoError(t, f.stakingKeeper.SetValidatorByConsAddr(f.Context(), val2))
	require.NoError(t, f.stakingKeeper.SetValidatorByConsAddr(f.Context(), val3))
	require.NoError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.Context(), val1))
	require.NoError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.Context(), val2))
	require.NoError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.Context(), val3))

	_, err = f.stakingKeeper.Delegate(f.Context(), addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.Context(), powers[0]), sdkstakingtypes.Unbonded, val1, true)
	require.NoError(t, err)
	_, _ = f.stakingKeeper.Delegate(f.Context(), addrs[1], f.stakingKeeper.TokensFromConsensusPower(f.Context(), powers[1]), sdkstakingtypes.Unbonded, val2, true)
	require.NoError(t, err)
	_, _ = f.stakingKeeper.Delegate(f.Context(), addrs[2], f.stakingKeeper.TokensFromConsensusPower(f.Context(), powers[2]), sdkstakingtypes.Unbonded, val3, true)
	require.NoError(t, err)

	_, err = f.stakingKeeper.EndBlocker(f.Context())
	require.NoError(t, err)

	return addrs, valAddrs, valPks
}
