// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {SedaDataTypes} from "../../libraries/SedaDataTypes.sol";
import {IRequestHandler} from "../../interfaces/IRequestHandler.sol";

/// @title RequestHandler
/// @notice Implements the RequestHandlerBase for managing Seda protocol requests
abstract contract RequestHandlerBase is IRequestHandler {
    // ============ Types & State ============

    // Define a unique storage slot for RequestHandlerBase
    bytes32 private constant REQUEST_HANDLER_STORAGE_SLOT =
        keccak256(abi.encode(uint256(keccak256("seda.requesthandler.storage")) - 1)) & ~bytes32(uint256(0xff));

    /// @custom:storage-location erc7201:seda.requesthandler.storage
    struct RequestHandlerStorage {
        // Mapping of request IDs to Request structs
        mapping(bytes32 => SedaDataTypes.Request) requests;
    }

    // ============ External Functions ============

    /// @notice Derives a request ID from the given inputs
    /// @param inputs The request inputs
    /// @return The derived request ID
    function deriveRequestId(SedaDataTypes.RequestInputs calldata inputs) external pure returns (bytes32) {
        return SedaDataTypes.deriveRequestId(inputs);
    }

    // ============ Public Functions ============

    /// @inheritdoc IRequestHandler
    function postRequest(
        SedaDataTypes.RequestInputs calldata inputs
    ) public payable virtual override(IRequestHandler) returns (bytes32) {
        if (inputs.replicationFactor == 0) {
            revert InvalidReplicationFactor();
        }

        bytes32 requestId = SedaDataTypes.deriveRequestId(inputs);
        if (bytes(_requestHandlerStorage().requests[requestId].version).length != 0) {
            revert RequestAlreadyExists(requestId);
        }

        _requestHandlerStorage().requests[requestId] = SedaDataTypes.Request({
            version: SedaDataTypes.VERSION,
            execProgramId: inputs.execProgramId,
            execInputs: inputs.execInputs,
            execGasLimit: inputs.execGasLimit,
            tallyProgramId: inputs.tallyProgramId,
            tallyInputs: inputs.tallyInputs,
            tallyGasLimit: inputs.tallyGasLimit,
            replicationFactor: inputs.replicationFactor,
            consensusFilter: inputs.consensusFilter,
            gasPrice: inputs.gasPrice,
            memo: inputs.memo
        });

        emit RequestPosted(requestId);
        return requestId;
    }

    /// @inheritdoc IRequestHandler
    function getRequest(
        bytes32 requestId
    ) public view virtual override(IRequestHandler) returns (SedaDataTypes.Request memory) {
        SedaDataTypes.Request memory request = _requestHandlerStorage().requests[requestId];
        // Version field is always set
        if (bytes(request.version).length == 0) {
            revert RequestNotFound(requestId);
        }

        return _requestHandlerStorage().requests[requestId];
    }

    // ============ Internal Functions ============

    /// @notice Returns the storage struct for the contract
    /// @dev Uses ERC-7201 storage pattern to access the storage struct at a specific slot
    /// @return s The storage struct containing the contract's state variables
    function _requestHandlerStorage() internal pure returns (RequestHandlerStorage storage s) {
        bytes32 slot = REQUEST_HANDLER_STORAGE_SLOT;
        // solhint-disable-next-line no-inline-assembly
        assembly {
            s.slot := slot
        }
    }
}
