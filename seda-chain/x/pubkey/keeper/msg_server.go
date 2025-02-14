package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type msgServer struct {
	*Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddKey(goCtx context.Context, msg *types.MsgAddKey) (*types.MsgAddKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message.
	err := msg.Validate()
	if err != nil {
		return nil, err
	}
	err = utils.ValidateSEDAPubKeys(msg.IndexedPubKeys)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid SEDA keys: %s", err)
	}

	// Verify that the validator exists.
	valAddr, err := m.validatorAddressCodec.StringToBytes(msg.ValidatorAddr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	_, err = m.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, sdkerrors.ErrNotFound.Wrapf("validator not found %s", msg.ValidatorAddr)
	}

	// Store the public keys.
	err = m.StoreIndexedPubKeys(ctx, valAddr, msg.IndexedPubKeys)
	if err != nil {
		return nil, err
	}
	return &types.MsgAddKeyResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
