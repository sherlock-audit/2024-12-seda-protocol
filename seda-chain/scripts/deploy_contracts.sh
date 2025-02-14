#!/bin/bash
set -e
set -x

CHAIN_ID=seda-1-local
RPC_URL=http://127.0.0.1:26657
BIN=$(git rev-parse --show-toplevel)/build/sedad
CONTRACT_WASM=$(git rev-parse --show-toplevel)/testutil/testwasms/core_contract.wasm
VOTING_PERIOD=30 # seconds
DEV_ACCOUNT=$($BIN keys show satoshi --keyring-backend test -a) # for sending wasm-storage txs
VOTE_ACCOUNT=$($BIN keys show satoshi --keyring-backend test -a) # for sending vote txs


echo "Deploying core contract"

OUTPUT="$(
    $BIN tx wasm store $CONTRACT_WASM \
    --node $RPC_URL \
    --chain-id $CHAIN_ID \
    --from $DEV_ACCOUNT \
    --keyring-backend test \
    --gas-prices 100000000000aseda \
    --gas auto \
    --gas-adjustment 1.3 \
    -y --output json \
    | $BIN query wait-tx --node $RPC_URL --output json\
)"

CORE_CODE_ID=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

echo "Instantiating core contract on code id $CORE_CODE_ID"

OUTPUT=$($BIN tx wasm-storage submit-proposal instantiate-core-contract $CORE_CODE_ID \
    '{"token":"aseda", "owner": "'$DEV_ACCOUNT'", "chain_id":"'$CHAIN_ID'" }' \
    $(printf '%02x' $CORE_CODE_ID) \
    --admin $DEV_ACCOUNT \
    --label core$CORE_CODE_ID \
    --title 'Core Contract' --summary 'Instantiates and registers core contract' --deposit 10000000aseda \
    --node $RPC_URL \
    --chain-id $CHAIN_ID \
    --from $DEV_ACCOUNT \
    --keyring-backend test \
    --gas-prices 100000000000aseda \
    --gas auto \
    --gas-adjustment 1.5 \
    -y --output json \
    | $BIN query wait-tx --node $RPC_URL --output json\
)

PROPOSAL_ID="$(echo "$OUTPUT" | jq '.events[] | select(.type == "submit_proposal") | .attributes[] | select(.key == "proposal_id") | .value' | sed 's/^"\(.*\)"$/\1/')"
$BIN tx gov vote $PROPOSAL_ID yes \
    --node $RPC_URL \
    --chain-id $CHAIN_ID \
    --from $VOTE_ACCOUNT \
    --keyring-backend test \
    --gas-prices 100000000000aseda \
    --gas auto \
    --gas-adjustment 1.6 \
    -y \
    | $BIN query wait-tx --node $RPC_URL --output json

sleep $VOTING_PERIOD

CORE_CONTRACT_ADDRESS=$($BIN query wasm-storage core-contract-registry --output json | jq -r '.address')
echo "Deployed core contract to: $CORE_CONTRACT_ADDRESS"

echo "Storing sample tally wasm"
$BIN tx wasm-storage store-oracle-program ./x/wasm-storage/keeper/testdata/sample_tally.wasm --from $DEV_ACCOUNT --keyring-backend test --gas auto --gas-adjustment 1.5 --gas-prices 10000000000aseda --chain-id $CHAIN_ID -y
