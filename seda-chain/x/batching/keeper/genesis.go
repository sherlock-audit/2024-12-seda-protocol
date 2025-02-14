package keeper

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// InitGenesis puts all data from genesis state into store.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.setCurrentBatchNum(ctx, data.CurrentBatchNumber); err != nil {
		panic(err)
	}
	for _, batch := range data.Batches {
		if err := k.setBatch(ctx, batch); err != nil {
			panic(err)
		}
	}
	for _, data := range data.BatchData {
		if err := k.setDataResultTreeEntry(ctx, data.BatchNumber, data.DataResultEntries); err != nil {
			panic(err)
		}
		for _, valEntry := range data.ValidatorEntries {
			if err := k.setValidatorTreeEntry(ctx, data.BatchNumber, valEntry); err != nil {
				panic(err)
			}
		}
		for _, sig := range data.BatchSignatures {
			if err := k.SetBatchSigSecp256k1(ctx, data.BatchNumber, sig.ValidatorAddress, sig.Secp256K1Signature); err != nil {
				panic(err)
			}
		}
	}
	for _, dataResult := range data.DataResults {
		if err := k.dataResults.Set(ctx, collections.Join3(false, dataResult.DataResult.DrId, dataResult.DataResult.DrBlockHeight), dataResult.DataResult); err != nil {
			panic(err)
		}
	}
	for _, batchAssignment := range data.BatchAssignments {
		if err := k.SetBatchAssignment(ctx, batchAssignment.DataRequestId, batchAssignment.DataRequestHeight, batchAssignment.BatchNumber); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	dataResults, err := k.getAllGenesisDataResults(ctx)
	if err != nil {
		panic(err)
	}
	batchAssignments, err := k.getAllBatchAssignments(ctx)
	if err != nil {
		panic(err)
	}
	curBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		panic(err)
	}
	batches, err := k.GetAllBatches(ctx)
	if err != nil {
		panic(err)
	}
	batchData := make([]types.BatchData, len(batches))
	for i, batch := range batches {
		data, err := k.GetBatchData(ctx, batch.BatchNumber)
		if err != nil {
			panic(err)
		}
		batchData[i] = data
	}
	return types.NewGenesisState(curBatchNum, batches, batchData, dataResults, batchAssignments)
}
