package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/batching"
	"github.com/sedaprotocol/seda-chain/x/batching/keeper"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx         sdk.Context
	keeper      *keeper.Keeper
	cdc         codec.Codec
	queryClient batchingtypes.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupSuite() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	config.Seal()
}

func (s *KeeperTestSuite) SetupTest() {
	batchingKeeper, enCfg, ctx := setupKeeper(s.T())
	s.keeper = batchingKeeper
	s.ctx = ctx
	s.cdc = enCfg.Codec

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enCfg.InterfaceRegistry)
	querier := keeper.NewQuerierImpl(*s.keeper)
	batchingtypes.RegisterQueryServer(queryHelper, querier)
	s.queryClient = batchingtypes.NewQueryClient(queryHelper)
}

func setupKeeper(t *testing.T) (*keeper.Keeper, moduletestutil.TestEncodingConfig, sdk.Context) {
	t.Helper()
	key := storetypes.NewKVStoreKey(batchingtypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(batching.AppModuleBasic{})

	keeper := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), nil, nil, nil, nil, nil, nil, nil)

	return &keeper, encCfg, ctx
}

func (s *KeeperTestSuite) TestKeeper_DataResult() {
	s.SetupTest()

	batchNum := uint64(rand.Intn(100) + 1)
	gasUsed := math.NewInt(20)
	mockDataResult := batchingtypes.DataResult{
		Version:        "0.0.1",
		DrId:           "74d7e8c9a77b7b4777153a32fcdf2424489f24cd59d3043eb2a30be7bba48306",
		Consensus:      true,
		ExitCode:       0,
		Result:         []byte("Ghkvq84TmIuEmU1ClubNxBjVXi8df5QhiNQEC5T8V6w="),
		BlockHeight:    12345,
		DrBlockHeight:  12343,
		GasUsed:        &gasUsed,
		PaybackAddress: "",
		SedaPayload:    "",
	}

	err := s.keeper.SetDataResultForBatching(s.ctx, mockDataResult)
	s.Require().NoError(err)

	res, err := s.queryClient.DataResult(s.ctx, &batchingtypes.QueryDataResultRequest{
		DataRequestId: mockDataResult.DrId,
	})
	s.Require().NoError(err)
	s.Require().Nil(res.BatchAssignment)
	s.Require().Equal(&mockDataResult, res.DataResult)

	// Query the data result after batching.
	err = s.keeper.MarkDataResultAsBatched(s.ctx, mockDataResult, batchNum)
	s.Require().NoError(err)

	res, err = s.queryClient.DataResult(s.ctx, &batchingtypes.QueryDataResultRequest{
		DataRequestId: mockDataResult.DrId,
	})
	s.Require().NoError(err)
	s.Require().Equal(&mockDataResult, res.DataResult)
	s.Require().Equal(&batchingtypes.BatchAssignment{
		BatchNumber:       batchNum,
		DataRequestId:     mockDataResult.DrId,
		DataRequestHeight: mockDataResult.DrBlockHeight,
	}, res.BatchAssignment)

	// Resolve and batch another data result for the same data request ID.
	mockDataResult2 := mockDataResult
	mockDataResult2.Id = "ccf12276c43cc61e0f3c6ace3e66872eda5df5ec753525a7bddab6fa3407e927"
	mockDataResult2.DrBlockHeight = 54321

	err = s.keeper.SetDataResultForBatching(s.ctx, mockDataResult2)
	s.Require().NoError(err)
	err = s.keeper.MarkDataResultAsBatched(s.ctx, mockDataResult2, batchNum)
	s.Require().NoError(err)

	res, err = s.queryClient.DataResult(s.ctx, &batchingtypes.QueryDataResultRequest{
		DataRequestId: mockDataResult2.DrId,
	})
	s.Require().NoError(err)
	s.Require().Equal(&batchingtypes.BatchAssignment{
		BatchNumber:       batchNum,
		DataRequestId:     mockDataResult2.DrId,
		DataRequestHeight: mockDataResult2.DrBlockHeight,
	}, res.BatchAssignment)
	s.Require().Equal(&mockDataResult2, res.DataResult)

	// We should still be able to query the first data result.
	res, err = s.queryClient.DataResult(s.ctx, &batchingtypes.QueryDataResultRequest{
		DataRequestId:     mockDataResult.DrId,
		DataRequestHeight: mockDataResult.DrBlockHeight,
	})
	s.Require().NoError(err)
	s.Require().Equal(&mockDataResult, res.DataResult)
	s.Require().Equal(&batchingtypes.BatchAssignment{
		BatchNumber:       batchNum,
		DataRequestId:     mockDataResult.DrId,
		DataRequestHeight: mockDataResult.DrBlockHeight,
	}, res.BatchAssignment)
}
