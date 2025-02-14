// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";

import {IRequestHandler} from "../interfaces/IRequestHandler.sol";
import {IResultHandler} from "../interfaces/IResultHandler.sol";
import {ISedaCore} from "../interfaces/ISedaCore.sol";
import {RequestHandlerBase} from "../core/abstract/RequestHandlerBase.sol";
import {SedaDataTypes} from "../libraries/SedaDataTypes.sol";

/// @title SedaCorePermissioned
/// @notice Core contract for the Seda protocol with permissioned access, managing requests and results
/// @dev Implements RequestHandlerBase, IResultHandler, AccessControl, and Pausable functionalities
/// @dev WARNING: This is a permissioned version of the SEDA core contract, primarily intended for testing
///      and controlled environments. It should not be used in production without careful consideration
///      of the centralization risks introduced by the permissioning system.
contract SedaPermissioned is ISedaCore, RequestHandlerBase, AccessControl, Pausable {
    using EnumerableSet for EnumerableSet.Bytes32Set;

    // ============ Types & State ============

    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");

    uint16 public maxReplicationFactor;
    mapping(bytes32 => SedaDataTypes.Result) public results;
    EnumerableSet.Bytes32Set private pendingRequests;

    // ============ Errors ============

    error FeesNotImplemented();

    // ============ Constructor ============

    /// @notice Contract constructor
    /// @param relayers The initial list of relayer addresses to be granted the RELAYER_ROLE
    /// @param initialMaxReplicationFactor The initial maximum replication factor
    constructor(address[] memory relayers, uint16 initialMaxReplicationFactor) {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
        _setRoleAdmin(ADMIN_ROLE, ADMIN_ROLE);
        _setRoleAdmin(RELAYER_ROLE, ADMIN_ROLE);

        // Grant RELAYER_ROLE to each address in the _relayers array
        for (uint256 i = 0; i < relayers.length; i++) {
            _grantRole(RELAYER_ROLE, relayers[i]);
        }

        maxReplicationFactor = initialMaxReplicationFactor;
    }

    // ============ External Functions ============

    /// @notice Sets the maximum replication factor that can be used for requests
    /// @param newMaxReplicationFactor The new maximum replication factor
    function setMaxReplicationFactor(uint16 newMaxReplicationFactor) external onlyRole(ADMIN_ROLE) {
        maxReplicationFactor = newMaxReplicationFactor;
    }

    /// @notice Posts a result for a request
    /// @param result The result data
    function postResult(
        SedaDataTypes.Result calldata result,
        uint64,
        bytes32[] calldata
    ) external payable override(IResultHandler) onlyRole(RELAYER_ROLE) whenNotPaused returns (bytes32) {
        bytes32 resultId = SedaDataTypes.deriveResultId(result);
        if (results[result.drId].drId != bytes32(0)) {
            revert ResultAlreadyExists(resultId);
        }
        results[result.drId] = result;
        _removePendingRequest(result.drId);
        emit ResultPosted(resultId);
        return resultId;
    }

    /// @notice Retrieves a result by its ID
    /// @param requestId The unique identifier of the result
    /// @return The result data associated with the given ID
    function getResult(bytes32 requestId) external view override returns (SedaDataTypes.Result memory) {
        if (results[requestId].drId == bytes32(0)) {
            revert ResultNotFound(requestId);
        }

        return results[requestId];
    }

    /// @inheritdoc IResultHandler
    /// @dev This mock implementation does not use a prover contract, so it returns address(0)
    /// @return The zero address (address(0))
    function getSedaProver() external pure override(IResultHandler) returns (address) {
        return address(0);
    }

    // ============ Public Functions ============

    /// @inheritdoc IRequestHandler
    /// @notice Posts a new request without any fees
    /// @param inputs The request inputs containing the request parameters
    /// @return The unique identifier of the posted request
    function postRequest(
        SedaDataTypes.RequestInputs calldata inputs
    ) public payable override(RequestHandlerBase, IRequestHandler) returns (bytes32) {
        return postRequest(inputs, 0, 0, 0);
    }

    /// @inheritdoc ISedaCore
    /// @notice Posts a new request
    /// @param inputs The request inputs
    /// @return requestId The ID of the posted request
    function postRequest(
        SedaDataTypes.RequestInputs calldata inputs,
        uint256,
        uint256,
        uint256
    ) public payable override(ISedaCore) whenNotPaused returns (bytes32) {
        // Check if amount is greater than 0
        if (msg.value != 0) {
            revert FeesNotImplemented();
        }

        // Check max replication factor first
        if (inputs.replicationFactor > maxReplicationFactor) {
            revert InvalidReplicationFactor();
        }

        // Call parent implementation which handles the rest
        bytes32 requestId = super.postRequest(inputs);

        // Add to pending requests (unique to this implementation)
        _addPendingRequest(requestId);

        return requestId;
    }

    /// @inheritdoc ISedaCore
    /// @notice Retrieves a list of pending request IDs
    /// @param offset The starting index in the pendingRequests set
    /// @param limit The maximum number of request IDs to return
    /// @return An array of pending request IDs
    function getPendingRequests(uint256 offset, uint256 limit) public view returns (PendingRequest[] memory) {
        uint256 totalRequests = pendingRequests.length();
        if (offset >= totalRequests) {
            return new PendingRequest[](0);
        }

        uint256 actualLimit = (offset + limit > totalRequests) ? totalRequests - offset : limit;
        PendingRequest[] memory queriedPendingRequests = new PendingRequest[](actualLimit);
        for (uint256 i = 0; i < actualLimit; i++) {
            bytes32 requestId = pendingRequests.at(offset + i);

            queriedPendingRequests[i] = PendingRequest({
                id: requestId,
                request: getRequest(requestId),
                requestor: address(0),
                timestamp: 0,
                requestFee: 0,
                resultFee: 0,
                batchFee: 0
            });
        }

        return queriedPendingRequests;
    }

    /// @inheritdoc ISedaCore
    /// @dev This is a mock implementation that does nothing
    function increaseFees(bytes32, uint256, uint256, uint256) external payable override(ISedaCore) {
        revert FeesNotImplemented();
    }

    // ============ Admin Functions ============

    /// @notice Adds a relayer
    /// @param account The address of the relayer to add
    function addRelayer(address account) external onlyRole(ADMIN_ROLE) {
        grantRole(RELAYER_ROLE, account);
    }

    /// @notice Removes a relayer
    /// @param account The address of the relayer to remove
    function removeRelayer(address account) external onlyRole(ADMIN_ROLE) {
        revokeRole(RELAYER_ROLE, account);
    }

    /// @notice Pauses the contract
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /// @notice Unpauses the contract
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }

    // ============ Internal Functions ============

    /// @notice Adds a request ID to the pendingRequests set
    /// @param requestId The ID of the request to add
    function _addPendingRequest(bytes32 requestId) internal {
        pendingRequests.add(requestId);
    }

    /// @notice Removes a request ID from the pendingRequests set if it exists
    /// @param requestId The ID of the request to remove
    function _removePendingRequest(bytes32 requestId) internal {
        pendingRequests.remove(requestId);
    }
}
