// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title SedaDataTypes Library
/// @notice Contains data structures and utility functions for the SEDA protocol
library SedaDataTypes {
    string internal constant VERSION = "0.0.1";

    /// @notice Input parameters for creating a data request
    struct RequestInputs {
        /// Identifier of Execution WASM binary
        bytes32 execProgramId;
        /// Inputs for Execution WASM binary
        bytes execInputs;
        /// Maximum of gas units for DR Execution
        uint64 execGasLimit;
        /// Identifier of Tally WASM binary
        bytes32 tallyProgramId;
        /// Inputs for Tally WASM binary
        bytes tallyInputs;
        /// Maximum of gas units for DR Tally
        uint64 tallyGasLimit;
        /// Amount of required DR executors
        uint16 replicationFactor;
        /// Filter applied before tally execution
        bytes consensusFilter;
        /// Amount of SEDA tokens per gas unit
        uint128 gasPrice;
        /// Public info attached to DR
        bytes memo;
    }

    /// @notice Full data request structure
    struct Request {
        /// Semantic Version
        string version;
        // DR definition
        /// Identifier of Execution WASM binary
        bytes32 execProgramId;
        /// Inputs for Execution WASM binary
        bytes execInputs;
        /// Maximum of gas units for DR Execution
        uint64 execGasLimit;
        /// Identifier of Tally WASM binary
        bytes32 tallyProgramId;
        /// Inputs for Tally WASM binary
        bytes tallyInputs;
        /// Maximum of gas units for DR Tally
        uint64 tallyGasLimit;
        /// Amount of required DR executors
        uint16 replicationFactor;
        /// Filter to be applied before tally execution
        bytes consensusFilter;
        /// Amount of SEDA tokens per gas unit
        uint128 gasPrice;
        /// Public info attached to DR
        bytes memo;
    }

    /// @notice Result of a data request execution
    struct Result {
        /// Semantic Version
        string version;
        /// Data Request Identifier
        bytes32 drId;
        /// True or false whether the reveal results are in consensus or not (≥ 66%)
        bool consensus;
        /// Exit code of Tally WASM binary execution
        uint8 exitCode;
        /// Result from Tally WASM binary execution
        bytes result;
        /// Block Height at which data request was finalized
        uint64 blockHeight;
        /// The timestamp of the block the data result is included
        uint64 blockTimestamp;
        /// Gas used by the complete data request execution
        uint128 gasUsed;
        // Fields from Data Request Execution
        /// Payback address set by the relayer
        bytes paybackAddress;
        /// Payload set by SEDA Protocol (e.g. OEV-enabled data requests)
        bytes sedaPayload;
    }

    /// @notice Represents a batch of data request results
    /// @dev This struct is used in the batch verification process
    struct Batch {
        uint64 batchHeight;
        uint64 blockHeight;
        bytes32 validatorsRoot;
        bytes32 resultsRoot;
        bytes32 provingMetadata;
    }

    /// @notice Proof structure for validator verification
    /// @dev Used in the validator set verification process
    struct ValidatorProof {
        uint32 votingPower;
        address signer;
        bytes32[] merkleProof;
    }

    /// @notice Derives a unique batch ID from a Batch struct
    /// @param batch The Batch struct to derive the ID from
    /// @return The derived batch ID
    function deriveBatchId(Batch memory batch) internal pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    bytes8(batch.batchHeight),
                    bytes8(batch.blockHeight),
                    batch.validatorsRoot,
                    batch.resultsRoot,
                    batch.provingMetadata
                )
            );
    }

    /// @notice Derives a unique request ID from RequestInputs
    /// @param inputs The RequestInputs struct to derive the ID from
    /// @return The derived request ID
    function deriveRequestId(RequestInputs memory inputs) internal pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(bytes(SedaDataTypes.VERSION)),
                    inputs.execProgramId,
                    keccak256(inputs.execInputs),
                    bytes8(inputs.execGasLimit),
                    inputs.tallyProgramId,
                    keccak256(inputs.tallyInputs),
                    bytes8(inputs.tallyGasLimit),
                    bytes2(inputs.replicationFactor),
                    keccak256(inputs.consensusFilter),
                    bytes16(inputs.gasPrice),
                    keccak256(inputs.memo)
                )
            );
    }

    /// @notice Derives a unique result ID from a Result struct
    /// @param result The Result struct to derive the ID from
    /// @return The derived result ID
    function deriveResultId(Result memory result) internal pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(bytes(SedaDataTypes.VERSION)),
                    result.drId,
                    result.consensus ? bytes1(0x01) : bytes1(0x00),
                    bytes1(result.exitCode),
                    keccak256(result.result),
                    bytes8(result.blockHeight),
                    bytes8(result.blockTimestamp),
                    bytes16(result.gasUsed),
                    keccak256(result.paybackAddress),
                    keccak256(result.sedaPayload)
                )
            );
    }
}
