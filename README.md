# SEDA Protocol contest details

- Join [Sherlock Discord](https://discord.gg/MABEWyASkp)
- Submit findings using the **Issues** page in your private contest repo (label issues as **Medium** or **High**)
- [Read for more details](https://docs.sherlock.xyz/audits/watsons)

# Q&A

### Q: On what chains are the smart contracts going to be deployed?
The SEDA protocol involves two types of smart contracts with distinct scopes:

1. CosmWASM contracts: These implement the core business logic of the protocol and are deployed exclusively on the SEDA Chain.

2. EVM contracts: These facilitate interoperability between SEDA and other networks. They are deployed on the following EVM-compatible chains:
 - Ethereum
 - Base
 - Arbitrum
 - Optimism
 - Ink
___

### Q: If you are integrating tokens, are you allowing only whitelisted tokens to work with the codebase or any complying with the standard? Are they assumed to have certain properties, e.g. be non-reentrant? Are there any types of [weird tokens](https://github.com/d-xo/weird-erc20) you want to integrate?
The SEDA Network operates with the SEDA Chain’s native token (SEDA), which follows coin type 118 and has 18 decimals.

The EVM contracts interact with the chain’s native token, primarily ETH on Ethereum and equivalent native assets on other chains. No additional token standards are integrated at this stage.
___

### Q: Are there any limitations on values set by admins (or other roles) in the codebase, including restrictions on array lengths?
The SEDA protocol is designed for progressive decentralization, but in its early stages, a failsafe mechanism based on roles and admin controls ensures stability, security, and incident response. These controls will be phased out or permanently disabled as the network matures.

1. SEDA Chain
Governance follows the standard Cosmos SDK governance model, where on-chain proposals and community voting determine module parameter changes and network upgrades. There are no arbitrary admin-controlled limitations beyond what is defined by governance.
A key aspect for the SEDA protocol is that WASM module parameters specify which accounts are authorized to upload and instantiate CosmWasm contracts, ensuring controlled and secure deployment.

2. CosmWASM Core Contracts
The Core Contracts owner account can:
- Pause and unpause the contract.
- Configure staking parameters (minimum stake, eligibility, allowlist enforcement).
- Manage the allowlist of executors (overlay nodes executing data requests).
The owner account and authorized deployment accounts are assigned to a security group account, a multi-signature account for enhanced security.

3. EVM Contracts
The contracts follow a UUPS upgradeable proxy pattern, where an owner can perform contract upgrades.
The Prover contract and the Core contract are PausableUpgradeable, allowing the owner to pause and unpause contract operations.

___

### Q: Are there any limitations on values set by admins (or other roles) in protocols you integrate with, including restrictions on array lengths?
No.
___

### Q: Is the codebase expected to comply with any specific EIPs?
No.
___

### Q: Are there any off-chain mechanisms involved in the protocol (e.g., keeper bots, arbitrage bots, etc.)? We assume these mechanisms will not misbehave, delay, or go offline unless otherwise specified.
There are two off-chain components of the SEDA Network; the SEDA Overlay Network and the decentralized Solver Network. 

SEDA Overlay Network

The SEDA Overlay Network is a multi-party computational network built to host tens of thousands of independently operated nodes. The primary responsibility of The Overlay Network is to query data sources as defined by a protocol Oracle Program.
The Oracle Program configuration determines the number of nodes elected, allowing for a customizable replication factor that balances scalability and decentralization. This secret committee independently queries the specified data sources and submits their results using a commit-reveal scheme. This process maintains data integrity and decentralized data querying, while eliminating the need for a single source of truth.

At the time of this audit, the SEDA Overlay Network will be launched in phased roll-out with a subset of allowlisted professional validators. The Overlay Network will be upgraded in the future to a permissionless model, incorporating staking and slashing mechanisms to secure participation and incentivize honest behavior. In the future, when a Data Request is submitted on the SEDA Chain, a randomly selected subset of Overlay Nodes forms a secret committee tasked with executing the data request binaries stored in the Oracle Program. This selection process leverages cryptographic sortition to ensure fairness and unpredictability, enhancing decentralization and security.

Decentralized Solver Network 

The Solver Network is responsible for delivering data requests from the SEDA Prover Contract to the SEDA Chain for execution and returning the Data Request results back to the SEDA Prover Contract on the origin network.
To ensure tamper resistance, data results are batched, and each batch is signed by multiple validators using ECDSA signatures. The Solver submits the signed batch to the SEDA Prover Contract, which verifies its integrity and authenticity. When individual results are later submitted, they are verified against the corresponding batch using proofs of inclusion, ensuring they originate from a valid, previously verified batch.

___

### Q: What properties/invariants do you want to hold even if breaking them has a low/unknown impact?
SEDA Chain
- All active validators must have a registered Proving Public Key for batch signing. This requirement only applies once a proving scheme is activated, such as through a network upgrade proposal.

CosmWASM Contracts & EVM Contracts
- Both contracts lock tokens upon posting of a data request to ensure that all involved parties receive their designated allocations.
- Once a data request is resolved, the entire amount should be fully distributed, ensuring no excess tokens remain in the contract.
- When there are no ongoing data requests, the contract’s token balance should always be zero.

___

### Q: Please discuss any design choices you made.
SEDA Chain
- Gas Allocation: The gas allocation model is designed to incentivize honest behavior in a simple yet effective way. Overlay nodes report their gas usage, but to discourage inflated reporting, the lowest gas reporter receives additional rewards compared to others. This ensures that nodes are incentivized to use and report only the necessary amount of gas, preventing unnecessary overuse of network resources.
- Fee Distribution: To maintain integrity in fee distribution, the protocol uses tally filtering to identify and exclude outliers. Only nodes non-outliers receive rewards, while outliers are filtered out. If the tally filter is too aggressive, meaning no clear consensus is reached, all participants are rewarded to prevent honest nodes from being unfairly penalized. Additionally, in cases where execution occurs but fails due to errors (e.g., no consensus or invalid filter input), only 80% of rewards are distributed, and 20% is burned, ensuring that execution efforts are still compensated while discouraging repeated failures.
- Refunding Unused Gas: Since gas usage is unpredictable due to the non-deterministic nature of external data, users can set a gas limit for execution and tally phases. Any unused gas is refunded, as overlay nodes are already incentivized to report their gas usage accurately due to the gas allocation mechanism. This ensures that users are not overcharged for gas they do not use while keeping the network efficient.
- Data Proxy Fees: To simplify cost predictability for data providers, fees are denominated in SEDA tokens. This approach makes it easier for providers to estimate their earnings per request. However, because execution costs are incurred in gas, the protocol translates these fees into gas units using the data request gas price, ensuring consistency across gas allocation and fee distribution.
- Batch Signing: Batch signing is implemented using vote extensions in ABCI++, as it provides an efficient method without introducing significant breaking changes to Cosmos SDK modules. This design choice allowed for seamless integration without modifying the block data structure, ensuring compatibility and maintainability.
- Batching Proving Metadata: Although this field is currently unused, it is reserved for future proving schemes. For example, proving schemes with signature aggregation, which would require storing an aggregated public key alongside the corresponding voting power of participating validators. This flexibility allows the protocol to evolve without requiring significant breaking changes.
- Double Batch Signing: To maintain consistency with existing validator security mechanisms, double batch signing incurs the same penalty as double block signing. This ensures that validators are held accountable for signing multiple conflicting batches, reinforcing network security and trust.

CosmWasm Core Contracts:
- Executor Allowlist: The allowlist is part of the phased decentralization of the Overlay Network. While any executor is technically eligible, only reputed and vetted operators are expected to be added. Inactive or malicious actors can be removed at any time to maintain network integrity.
- Multiple Identities: Users do not need separate wallets for each identity they operate. While this feature is less relevant under the current allowlist model, it will become crucial once the Overlay Network is fully decentralized. It allows users to run multiple identities efficiently, maximizing hardware utilization for vertical scaling, which enhances network security and performance.
- Pausable: A failsafe mechanism to mitigate and respond to incidents or unexpected failures. This allows intervention when necessary to maintain system stability.
- No Slashing / No Unstaking: This is a temporary measure as part of the phased decentralization process. Slashing and unstaking mechanisms will be introduced once the Overlay Network becomes permissionless. Until then, inactive or malicious actors can still be removed from the allowlist to uphold network integrity.

EVM Contracts:
- Pausable: A failsafe mechanism designed to mitigate and respond to incidents. This ensures that contract operations can be temporarily halted if necessary to maintain system integrity.
- UUPS Proxy Pattern: The contracts follow the UUPS (Universal Upgradeable Proxy Standard) pattern, providing upgradeability in the early stages of development. This enhances developer experience while allowing improvements over time. The upgradeability feature can be renounced, ensuring the contracts remain immutable once they reach a stable state.
- Not Optimized Implementations: The contracts serve as reference implementations, prioritizing clarity and functionality over full optimization. While gas efficiency improvements are possible, the current design ensures broad compatibility with various use cases, even if storage and execution costs are not fully optimized.

Further details are available in the 'SEDA Protocol Audit Overview' document listed below.
___

### Q: Please provide links to previous audits (if any).
SEDA engaged Trail of Bits to review the security of its token migration contracts and the SEDA chain’s staking and vesting modules.
https://github.com/trailofbits/publications/blob/master/reviews/2024-03-seda-chaintokenmigration-securityreview.pdf

___

### Q: Please list any relevant protocol resources.
SEDA Protocol Audit Overview: https://sedaprotocol.notion.site/SEDA-Protocol-Audit-Overview-190a68d575ca807ca2a2d4e232a77781

SEDA Website: https://seda.xyz

Additional Resources
- SEDA Primer for Key Features: https://docs.seda.xyz/home/overview/seda-overview/seda-primer-for-key-features
- SEDA Network Architecture: https://docs.seda.xyz/home/overview/seda-network-architecture
- Data Requests: https://docs.seda.xyz/home/for-developers/data-requests
- Accessing Data from any Network: https://docs.seda.xyz/home/for-developers/access-data-from-any-network
- Building an Oracle Program: https://docs.seda.xyz/home/for-developers/building-an-oracle-program

___

### Q: Additional audit information.
The following areas are of particular importance for the audit:
- Data Integrity: While data integrity mechanisms rely on battle-tested cryptographic primitives, they are fundamental to the SEDA protocol’s correctness. Ensuring that results remain tamper-proof and resistant to manipulation is a top priority.
- Gas Allocation and Fee Payments: The gas allocation and fee distribution processes are inherently complex. Special attention should be given to potential edge cases that could impact fairness, efficiency, or execution reliability.
- Tally WASM VM Robustness: The robustness of the Tally WASM VM is critical to ensuring uninterrupted operations. Possible denial-of-service (DoS) vectors or unexpected behavior could lead to operational disruptions, affecting data request execution and overall system stability.



# Audit scope

[seda-chain @ 58347a2cac169e697b84dd3d4d28669a76503cc3](https://github.com/sedaprotocol/seda-chain/tree/58347a2cac169e697b84dd3d4d28669a76503cc3)
- [seda-chain/app/abci/errors.go](seda-chain/app/abci/errors.go)
- [seda-chain/app/abci/expected_keepers.go](seda-chain/app/abci/expected_keepers.go)
- [seda-chain/app/abci/handlers.go](seda-chain/app/abci/handlers.go)
- [seda-chain/app/ante.go](seda-chain/app/ante.go)
- [seda-chain/app/app.go](seda-chain/app/app.go)
- [seda-chain/app/encoding.go](seda-chain/app/encoding.go)
- [seda-chain/app/export.go](seda-chain/app/export.go)
- [seda-chain/app/genesis.go](seda-chain/app/genesis.go)
- [seda-chain/app/keepers/keepers.go](seda-chain/app/keepers/keepers.go)
- [seda-chain/app/modules.go](seda-chain/app/modules.go)
- [seda-chain/app/params/config.go](seda-chain/app/params/config.go)
- [seda-chain/app/params/encoding.go](seda-chain/app/params/encoding.go)
- [seda-chain/app/upgrades.go](seda-chain/app/upgrades.go)
- [seda-chain/app/upgrades/mainnet/v1/constants.go](seda-chain/app/upgrades/mainnet/v1/constants.go)
- [seda-chain/app/upgrades/types.go](seda-chain/app/upgrades/types.go)
- [seda-chain/app/utils/merkle.go](seda-chain/app/utils/merkle.go)
- [seda-chain/app/utils/merkle_proof.go](seda-chain/app/utils/merkle_proof.go)
- [seda-chain/app/utils/print_info.go](seda-chain/app/utils/print_info.go)
- [seda-chain/app/utils/seda_keys.go](seda-chain/app/utils/seda_keys.go)
- [seda-chain/app/utils/seda_signer.go](seda-chain/app/utils/seda_signer.go)
- [seda-chain/app/wasm.go](seda-chain/app/wasm.go)
- [seda-chain/cmd/sedad/cmd/config_defaults.go](seda-chain/cmd/sedad/cmd/config_defaults.go)
- [seda-chain/cmd/sedad/cmd/genaccounts.go](seda-chain/cmd/sedad/cmd/genaccounts.go)
- [seda-chain/cmd/sedad/cmd/git_download.go](seda-chain/cmd/sedad/cmd/git_download.go)
- [seda-chain/cmd/sedad/cmd/init.go](seda-chain/cmd/sedad/cmd/init.go)
- [seda-chain/cmd/sedad/cmd/init_cmds.go](seda-chain/cmd/sedad/cmd/init_cmds.go)
- [seda-chain/cmd/sedad/cmd/root.go](seda-chain/cmd/sedad/cmd/root.go)
- [seda-chain/cmd/sedad/cmd/rosetta_disabled.go](seda-chain/cmd/sedad/cmd/rosetta_disabled.go)
- [seda-chain/cmd/sedad/cmd/rosetta_enabled.go](seda-chain/cmd/sedad/cmd/rosetta_enabled.go)
- [seda-chain/cmd/sedad/gentx/collect_gentxs.go](seda-chain/cmd/sedad/gentx/collect_gentxs.go)
- [seda-chain/cmd/sedad/gentx/gentx.go](seda-chain/cmd/sedad/gentx/gentx.go)
- [seda-chain/cmd/sedad/main.go](seda-chain/cmd/sedad/main.go)
- [seda-chain/go.mod](seda-chain/go.mod)
- [seda-chain/go.sum](seda-chain/go.sum)
- [seda-chain/proto/sedachain/batching/v1/batching.proto](seda-chain/proto/sedachain/batching/v1/batching.proto)
- [seda-chain/proto/sedachain/batching/v1/evidence.proto](seda-chain/proto/sedachain/batching/v1/evidence.proto)
- [seda-chain/proto/sedachain/batching/v1/genesis.proto](seda-chain/proto/sedachain/batching/v1/genesis.proto)
- [seda-chain/proto/sedachain/batching/v1/query.proto](seda-chain/proto/sedachain/batching/v1/query.proto)
- [seda-chain/proto/sedachain/data_proxy/v1/data_proxy.proto](seda-chain/proto/sedachain/data_proxy/v1/data_proxy.proto)
- [seda-chain/proto/sedachain/data_proxy/v1/genesis.proto](seda-chain/proto/sedachain/data_proxy/v1/genesis.proto)
- [seda-chain/proto/sedachain/data_proxy/v1/query.proto](seda-chain/proto/sedachain/data_proxy/v1/query.proto)
- [seda-chain/proto/sedachain/data_proxy/v1/tx.proto](seda-chain/proto/sedachain/data_proxy/v1/tx.proto)
- [seda-chain/proto/sedachain/pubkey/v1/genesis.proto](seda-chain/proto/sedachain/pubkey/v1/genesis.proto)
- [seda-chain/proto/sedachain/pubkey/v1/pubkey.proto](seda-chain/proto/sedachain/pubkey/v1/pubkey.proto)
- [seda-chain/proto/sedachain/pubkey/v1/query.proto](seda-chain/proto/sedachain/pubkey/v1/query.proto)
- [seda-chain/proto/sedachain/pubkey/v1/tx.proto](seda-chain/proto/sedachain/pubkey/v1/tx.proto)
- [seda-chain/proto/sedachain/staking/v1/tx.proto](seda-chain/proto/sedachain/staking/v1/tx.proto)
- [seda-chain/proto/sedachain/tally/v1/genesis.proto](seda-chain/proto/sedachain/tally/v1/genesis.proto)
- [seda-chain/proto/sedachain/tally/v1/query.proto](seda-chain/proto/sedachain/tally/v1/query.proto)
- [seda-chain/proto/sedachain/tally/v1/tally.proto](seda-chain/proto/sedachain/tally/v1/tally.proto)
- [seda-chain/proto/sedachain/tally/v1/tx.proto](seda-chain/proto/sedachain/tally/v1/tx.proto)
- [seda-chain/proto/sedachain/vesting/v1/tx.proto](seda-chain/proto/sedachain/vesting/v1/tx.proto)
- [seda-chain/proto/sedachain/vesting/v1/vesting.proto](seda-chain/proto/sedachain/vesting/v1/vesting.proto)
- [seda-chain/proto/sedachain/wasm_storage/v1/genesis.proto](seda-chain/proto/sedachain/wasm_storage/v1/genesis.proto)
- [seda-chain/proto/sedachain/wasm_storage/v1/query.proto](seda-chain/proto/sedachain/wasm_storage/v1/query.proto)
- [seda-chain/proto/sedachain/wasm_storage/v1/tx.proto](seda-chain/proto/sedachain/wasm_storage/v1/tx.proto)
- [seda-chain/proto/sedachain/wasm_storage/v1/wasm_storage.proto](seda-chain/proto/sedachain/wasm_storage/v1/wasm_storage.proto)
- [seda-chain/tools/tools.go](seda-chain/tools/tools.go)
- [seda-chain/x/batching/client/cli/query.go](seda-chain/x/batching/client/cli/query.go)
- [seda-chain/x/batching/client/cli/tx.go](seda-chain/x/batching/client/cli/tx.go)
- [seda-chain/x/batching/keeper/batch.go](seda-chain/x/batching/keeper/batch.go)
- [seda-chain/x/batching/keeper/data_result.go](seda-chain/x/batching/keeper/data_result.go)
- [seda-chain/x/batching/keeper/endblock.go](seda-chain/x/batching/keeper/endblock.go)
- [seda-chain/x/batching/keeper/evidence.go](seda-chain/x/batching/keeper/evidence.go)
- [seda-chain/x/batching/keeper/genesis.go](seda-chain/x/batching/keeper/genesis.go)
- [seda-chain/x/batching/keeper/keeper.go](seda-chain/x/batching/keeper/keeper.go)
- [seda-chain/x/batching/keeper/querier.go](seda-chain/x/batching/keeper/querier.go)
- [seda-chain/x/batching/module.go](seda-chain/x/batching/module.go)
- [seda-chain/x/batching/types/batch.go](seda-chain/x/batching/types/batch.go)
- [seda-chain/x/batching/types/codec.go](seda-chain/x/batching/types/codec.go)
- [seda-chain/x/batching/types/data_result.go](seda-chain/x/batching/types/data_result.go)
- [seda-chain/x/batching/types/errors.go](seda-chain/x/batching/types/errors.go)
- [seda-chain/x/batching/types/events.go](seda-chain/x/batching/types/events.go)
- [seda-chain/x/batching/types/evidence.go](seda-chain/x/batching/types/evidence.go)
- [seda-chain/x/batching/types/expected_keepers.go](seda-chain/x/batching/types/expected_keepers.go)
- [seda-chain/x/batching/types/genesis.go](seda-chain/x/batching/types/genesis.go)
- [seda-chain/x/batching/types/keys.go](seda-chain/x/batching/types/keys.go)
- [seda-chain/x/data-proxy/client/cli/query.go](seda-chain/x/data-proxy/client/cli/query.go)
- [seda-chain/x/data-proxy/client/cli/tx.go](seda-chain/x/data-proxy/client/cli/tx.go)
- [seda-chain/x/data-proxy/keeper/abci.go](seda-chain/x/data-proxy/keeper/abci.go)
- [seda-chain/x/data-proxy/keeper/genesis.go](seda-chain/x/data-proxy/keeper/genesis.go)
- [seda-chain/x/data-proxy/keeper/grpc_query.go](seda-chain/x/data-proxy/keeper/grpc_query.go)
- [seda-chain/x/data-proxy/keeper/keeper.go](seda-chain/x/data-proxy/keeper/keeper.go)
- [seda-chain/x/data-proxy/keeper/msg_server.go](seda-chain/x/data-proxy/keeper/msg_server.go)
- [seda-chain/x/data-proxy/keeper/params.go](seda-chain/x/data-proxy/keeper/params.go)
- [seda-chain/x/data-proxy/module.go](seda-chain/x/data-proxy/module.go)
- [seda-chain/x/data-proxy/types/codec.go](seda-chain/x/data-proxy/types/codec.go)
- [seda-chain/x/data-proxy/types/errors.go](seda-chain/x/data-proxy/types/errors.go)
- [seda-chain/x/data-proxy/types/events.go](seda-chain/x/data-proxy/types/events.go)
- [seda-chain/x/data-proxy/types/genesis.go](seda-chain/x/data-proxy/types/genesis.go)
- [seda-chain/x/data-proxy/types/keys.go](seda-chain/x/data-proxy/types/keys.go)
- [seda-chain/x/data-proxy/types/params.go](seda-chain/x/data-proxy/types/params.go)
- [seda-chain/x/data-proxy/types/proxy_config.go](seda-chain/x/data-proxy/types/proxy_config.go)
- [seda-chain/x/data-proxy/types/tx.go](seda-chain/x/data-proxy/types/tx.go)
- [seda-chain/x/pubkey/client/cli/keyfile.go](seda-chain/x/pubkey/client/cli/keyfile.go)
- [seda-chain/x/pubkey/client/cli/query.go](seda-chain/x/pubkey/client/cli/query.go)
- [seda-chain/x/pubkey/client/cli/tx.go](seda-chain/x/pubkey/client/cli/tx.go)
- [seda-chain/x/pubkey/keeper/endblock.go](seda-chain/x/pubkey/keeper/endblock.go)
- [seda-chain/x/pubkey/keeper/genesis.go](seda-chain/x/pubkey/keeper/genesis.go)
- [seda-chain/x/pubkey/keeper/grpc_query.go](seda-chain/x/pubkey/keeper/grpc_query.go)
- [seda-chain/x/pubkey/keeper/keeper.go](seda-chain/x/pubkey/keeper/keeper.go)
- [seda-chain/x/pubkey/keeper/msg_server.go](seda-chain/x/pubkey/keeper/msg_server.go)
- [seda-chain/x/pubkey/module.go](seda-chain/x/pubkey/module.go)
- [seda-chain/x/pubkey/types/codec.go](seda-chain/x/pubkey/types/codec.go)
- [seda-chain/x/pubkey/types/events.go](seda-chain/x/pubkey/types/events.go)
- [seda-chain/x/pubkey/types/expected_keepers.go](seda-chain/x/pubkey/types/expected_keepers.go)
- [seda-chain/x/pubkey/types/genesis.go](seda-chain/x/pubkey/types/genesis.go)
- [seda-chain/x/pubkey/types/keys.go](seda-chain/x/pubkey/types/keys.go)
- [seda-chain/x/pubkey/types/params.go](seda-chain/x/pubkey/types/params.go)
- [seda-chain/x/pubkey/types/tx.go](seda-chain/x/pubkey/types/tx.go)
- [seda-chain/x/slashing/module.go](seda-chain/x/slashing/module.go)
- [seda-chain/x/slashing/msg_server.go](seda-chain/x/slashing/msg_server.go)
- [seda-chain/x/staking/autocli.go](seda-chain/x/staking/autocli.go)
- [seda-chain/x/staking/client/cli/tx.go](seda-chain/x/staking/client/cli/tx.go)
- [seda-chain/x/staking/client/cli/utils.go](seda-chain/x/staking/client/cli/utils.go)
- [seda-chain/x/staking/keeper/invariants.go](seda-chain/x/staking/keeper/invariants.go)
- [seda-chain/x/staking/keeper/keeper.go](seda-chain/x/staking/keeper/keeper.go)
- [seda-chain/x/staking/keeper/msg_server.go](seda-chain/x/staking/keeper/msg_server.go)
- [seda-chain/x/staking/module.go](seda-chain/x/staking/module.go)
- [seda-chain/x/staking/types/codec.go](seda-chain/x/staking/types/codec.go)
- [seda-chain/x/staking/types/expected_keepers.go](seda-chain/x/staking/types/expected_keepers.go)
- [seda-chain/x/staking/types/msg.go](seda-chain/x/staking/types/msg.go)
- [seda-chain/x/tally/client/cli/query.go](seda-chain/x/tally/client/cli/query.go)
- [seda-chain/x/tally/keeper/endblock.go](seda-chain/x/tally/keeper/endblock.go)
- [seda-chain/x/tally/keeper/filter.go](seda-chain/x/tally/keeper/filter.go)
- [seda-chain/x/tally/keeper/gas_meter.go](seda-chain/x/tally/keeper/gas_meter.go)
- [seda-chain/x/tally/keeper/genesis.go](seda-chain/x/tally/keeper/genesis.go)
- [seda-chain/x/tally/keeper/grpc_query.go](seda-chain/x/tally/keeper/grpc_query.go)
- [seda-chain/x/tally/keeper/keeper.go](seda-chain/x/tally/keeper/keeper.go)
- [seda-chain/x/tally/keeper/msg_server.go](seda-chain/x/tally/keeper/msg_server.go)
- [seda-chain/x/tally/keeper/tally_vm.go](seda-chain/x/tally/keeper/tally_vm.go)
- [seda-chain/x/tally/module.go](seda-chain/x/tally/module.go)
- [seda-chain/x/tally/types/abci_types.go](seda-chain/x/tally/types/abci_types.go)
- [seda-chain/x/tally/types/codec.go](seda-chain/x/tally/types/codec.go)
- [seda-chain/x/tally/types/data_result.go](seda-chain/x/tally/types/data_result.go)
- [seda-chain/x/tally/types/errors.go](seda-chain/x/tally/types/errors.go)
- [seda-chain/x/tally/types/events.go](seda-chain/x/tally/types/events.go)
- [seda-chain/x/tally/types/expected_keepers.go](seda-chain/x/tally/types/expected_keepers.go)
- [seda-chain/x/tally/types/filters.go](seda-chain/x/tally/types/filters.go)
- [seda-chain/x/tally/types/filters_util.go](seda-chain/x/tally/types/filters_util.go)
- [seda-chain/x/tally/types/gas_meter.go](seda-chain/x/tally/types/gas_meter.go)
- [seda-chain/x/tally/types/genesis.go](seda-chain/x/tally/types/genesis.go)
- [seda-chain/x/tally/types/keys.go](seda-chain/x/tally/types/keys.go)
- [seda-chain/x/tally/types/params.go](seda-chain/x/tally/types/params.go)
- [seda-chain/x/tally/types/sigma_multiplier.go](seda-chain/x/tally/types/sigma_multiplier.go)
- [seda-chain/x/tally/types/telemetry.go](seda-chain/x/tally/types/telemetry.go)
- [seda-chain/x/tally/types/types.go](seda-chain/x/tally/types/types.go)
- [seda-chain/x/tally/types/vm_exit_codes.go](seda-chain/x/tally/types/vm_exit_codes.go)
- [seda-chain/x/vesting/client/cli/tx.go](seda-chain/x/vesting/client/cli/tx.go)
- [seda-chain/x/vesting/keeper/msg_server.go](seda-chain/x/vesting/keeper/msg_server.go)
- [seda-chain/x/vesting/module.go](seda-chain/x/vesting/module.go)
- [seda-chain/x/vesting/types/codec.go](seda-chain/x/vesting/types/codec.go)
- [seda-chain/x/vesting/types/errors.go](seda-chain/x/vesting/types/errors.go)
- [seda-chain/x/vesting/types/expected_keepers.go](seda-chain/x/vesting/types/expected_keepers.go)
- [seda-chain/x/vesting/types/msgs.go](seda-chain/x/vesting/types/msgs.go)
- [seda-chain/x/vesting/types/types.go](seda-chain/x/vesting/types/types.go)
- [seda-chain/x/vesting/types/vesting.go](seda-chain/x/vesting/types/vesting.go)
- [seda-chain/x/wasm-storage/client/cli/decoder.go](seda-chain/x/wasm-storage/client/cli/decoder.go)
- [seda-chain/x/wasm-storage/client/cli/gov_tx.go](seda-chain/x/wasm-storage/client/cli/gov_tx.go)
- [seda-chain/x/wasm-storage/client/cli/query.go](seda-chain/x/wasm-storage/client/cli/query.go)
- [seda-chain/x/wasm-storage/client/cli/tx.go](seda-chain/x/wasm-storage/client/cli/tx.go)
- [seda-chain/x/wasm-storage/keeper/genesis.go](seda-chain/x/wasm-storage/keeper/genesis.go)
- [seda-chain/x/wasm-storage/keeper/keeper.go](seda-chain/x/wasm-storage/keeper/keeper.go)
- [seda-chain/x/wasm-storage/keeper/msg_server.go](seda-chain/x/wasm-storage/keeper/msg_server.go)
- [seda-chain/x/wasm-storage/keeper/querier.go](seda-chain/x/wasm-storage/keeper/querier.go)
- [seda-chain/x/wasm-storage/module.go](seda-chain/x/wasm-storage/module.go)
- [seda-chain/x/wasm-storage/types/codec.go](seda-chain/x/wasm-storage/types/codec.go)
- [seda-chain/x/wasm-storage/types/errors.go](seda-chain/x/wasm-storage/types/errors.go)
- [seda-chain/x/wasm-storage/types/events.go](seda-chain/x/wasm-storage/types/events.go)
- [seda-chain/x/wasm-storage/types/expected_keepers.go](seda-chain/x/wasm-storage/types/expected_keepers.go)
- [seda-chain/x/wasm-storage/types/genesis.go](seda-chain/x/wasm-storage/types/genesis.go)
- [seda-chain/x/wasm-storage/types/keys.go](seda-chain/x/wasm-storage/types/keys.go)
- [seda-chain/x/wasm-storage/types/msgs.go](seda-chain/x/wasm-storage/types/msgs.go)
- [seda-chain/x/wasm-storage/types/params.go](seda-chain/x/wasm-storage/types/params.go)
- [seda-chain/x/wasm-storage/types/wasm.go](seda-chain/x/wasm-storage/types/wasm.go)

[seda-chain-contracts @ 9ab67eda2bd67e728f1bbf8379c9581431dee21d](https://github.com/sedaprotocol/seda-chain-contracts/tree/9ab67eda2bd67e728f1bbf8379c9581431dee21d)
- [seda-chain-contracts/contract/src/consts.rs](seda-chain-contracts/contract/src/consts.rs)
- [seda-chain-contracts/contract/src/contract.rs](seda-chain-contracts/contract/src/contract.rs)
- [seda-chain-contracts/contract/src/error.rs](seda-chain-contracts/contract/src/error.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/execute/commit_result.rs](seda-chain-contracts/contract/src/msgs/data_requests/execute/commit_result.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/execute/dr_events.rs](seda-chain-contracts/contract/src/msgs/data_requests/execute/dr_events.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/execute/mod.rs](seda-chain-contracts/contract/src/msgs/data_requests/execute/mod.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/execute/post_request.rs](seda-chain-contracts/contract/src/msgs/data_requests/execute/post_request.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/execute/reveal_result.rs](seda-chain-contracts/contract/src/msgs/data_requests/execute/reveal_result.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/execute/set_timeout_config.rs](seda-chain-contracts/contract/src/msgs/data_requests/execute/set_timeout_config.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/mod.rs](seda-chain-contracts/contract/src/msgs/data_requests/mod.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/query.rs](seda-chain-contracts/contract/src/msgs/data_requests/query.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/state/data_requests_map.rs](seda-chain-contracts/contract/src/msgs/data_requests/state/data_requests_map.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/state/mod.rs](seda-chain-contracts/contract/src/msgs/data_requests/state/mod.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/state/timeouts.rs](seda-chain-contracts/contract/src/msgs/data_requests/state/timeouts.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/sudo/expire_data_requests.rs](seda-chain-contracts/contract/src/msgs/data_requests/sudo/expire_data_requests.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/sudo/mod.rs](seda-chain-contracts/contract/src/msgs/data_requests/sudo/mod.rs)
- [seda-chain-contracts/contract/src/msgs/data_requests/sudo/remove_requests.rs](seda-chain-contracts/contract/src/msgs/data_requests/sudo/remove_requests.rs)
- [seda-chain-contracts/contract/src/msgs/enumerable_set.rs](seda-chain-contracts/contract/src/msgs/enumerable_set.rs)
- [seda-chain-contracts/contract/src/msgs/mod.rs](seda-chain-contracts/contract/src/msgs/mod.rs)
- [seda-chain-contracts/contract/src/msgs/owner/execute/accept_ownership.rs](seda-chain-contracts/contract/src/msgs/owner/execute/accept_ownership.rs)
- [seda-chain-contracts/contract/src/msgs/owner/execute/add_to_allowlist.rs](seda-chain-contracts/contract/src/msgs/owner/execute/add_to_allowlist.rs)
- [seda-chain-contracts/contract/src/msgs/owner/execute/mod.rs](seda-chain-contracts/contract/src/msgs/owner/execute/mod.rs)
- [seda-chain-contracts/contract/src/msgs/owner/execute/pause.rs](seda-chain-contracts/contract/src/msgs/owner/execute/pause.rs)
- [seda-chain-contracts/contract/src/msgs/owner/execute/remove_from_allowlist.rs](seda-chain-contracts/contract/src/msgs/owner/execute/remove_from_allowlist.rs)
- [seda-chain-contracts/contract/src/msgs/owner/execute/transfer_ownership.rs](seda-chain-contracts/contract/src/msgs/owner/execute/transfer_ownership.rs)
- [seda-chain-contracts/contract/src/msgs/owner/execute/unpause.rs](seda-chain-contracts/contract/src/msgs/owner/execute/unpause.rs)
- [seda-chain-contracts/contract/src/msgs/owner/mod.rs](seda-chain-contracts/contract/src/msgs/owner/mod.rs)
- [seda-chain-contracts/contract/src/msgs/owner/query.rs](seda-chain-contracts/contract/src/msgs/owner/query.rs)
- [seda-chain-contracts/contract/src/msgs/owner/state.rs](seda-chain-contracts/contract/src/msgs/owner/state.rs)
- [seda-chain-contracts/contract/src/msgs/owner/utils.rs](seda-chain-contracts/contract/src/msgs/owner/utils.rs)
- [seda-chain-contracts/contract/src/msgs/staking/execute/mod.rs](seda-chain-contracts/contract/src/msgs/staking/execute/mod.rs)
- [seda-chain-contracts/contract/src/msgs/staking/execute/set_staking_config.rs](seda-chain-contracts/contract/src/msgs/staking/execute/set_staking_config.rs)
- [seda-chain-contracts/contract/src/msgs/staking/execute/stake.rs](seda-chain-contracts/contract/src/msgs/staking/execute/stake.rs)
- [seda-chain-contracts/contract/src/msgs/staking/execute/staking_events.rs](seda-chain-contracts/contract/src/msgs/staking/execute/staking_events.rs)
- [seda-chain-contracts/contract/src/msgs/staking/execute/unstake.rs](seda-chain-contracts/contract/src/msgs/staking/execute/unstake.rs)
- [seda-chain-contracts/contract/src/msgs/staking/execute/withdraw.rs](seda-chain-contracts/contract/src/msgs/staking/execute/withdraw.rs)
- [seda-chain-contracts/contract/src/msgs/staking/mod.rs](seda-chain-contracts/contract/src/msgs/staking/mod.rs)
- [seda-chain-contracts/contract/src/msgs/staking/query.rs](seda-chain-contracts/contract/src/msgs/staking/query.rs)
- [seda-chain-contracts/contract/src/msgs/staking/state/is_eligible_for_dr.rs](seda-chain-contracts/contract/src/msgs/staking/state/is_eligible_for_dr.rs)
- [seda-chain-contracts/contract/src/msgs/staking/state/mod.rs](seda-chain-contracts/contract/src/msgs/staking/state/mod.rs)
- [seda-chain-contracts/contract/src/msgs/staking/state/stakers_map.rs](seda-chain-contracts/contract/src/msgs/staking/state/stakers_map.rs)
- [seda-chain-contracts/contract/src/state.rs](seda-chain-contracts/contract/src/state.rs)
- [seda-chain-contracts/contract/src/types.rs](seda-chain-contracts/contract/src/types.rs)
- [seda-chain-contracts/contract/src/utils.rs](seda-chain-contracts/contract/src/utils.rs)

[seda-evm-contracts @ 07f329125cb62eacce32f5057ba8e59fc4c67f7d](https://github.com/sedaprotocol/seda-evm-contracts/tree/07f329125cb62eacce32f5057ba8e59fc4c67f7d)
- [seda-evm-contracts/contracts/core/SedaCoreV1.sol](seda-evm-contracts/contracts/core/SedaCoreV1.sol)
- [seda-evm-contracts/contracts/core/abstract/RequestHandlerBase.sol](seda-evm-contracts/contracts/core/abstract/RequestHandlerBase.sol)
- [seda-evm-contracts/contracts/core/abstract/ResultHandlerBase.sol](seda-evm-contracts/contracts/core/abstract/ResultHandlerBase.sol)
- [seda-evm-contracts/contracts/interfaces/IProver.sol](seda-evm-contracts/contracts/interfaces/IProver.sol)
- [seda-evm-contracts/contracts/interfaces/IRequestHandler.sol](seda-evm-contracts/contracts/interfaces/IRequestHandler.sol)
- [seda-evm-contracts/contracts/interfaces/IResultHandler.sol](seda-evm-contracts/contracts/interfaces/IResultHandler.sol)
- [seda-evm-contracts/contracts/interfaces/ISedaCore.sol](seda-evm-contracts/contracts/interfaces/ISedaCore.sol)
- [seda-evm-contracts/contracts/libraries/SedaDataTypes.sol](seda-evm-contracts/contracts/libraries/SedaDataTypes.sol)
- [seda-evm-contracts/contracts/provers/Secp256k1ProverV1.sol](seda-evm-contracts/contracts/provers/Secp256k1ProverV1.sol)
- [seda-evm-contracts/contracts/provers/abstract/ProverBase.sol](seda-evm-contracts/contracts/provers/abstract/ProverBase.sol)

[seda-wasm-vm @ 8a0a7274119150b72c02cedce3edcc16b43d1599](https://github.com/sedaprotocol/seda-wasm-vm/tree/8a0a7274119150b72c02cedce3edcc16b43d1599)
- [seda-wasm-vm/libtallyvm/build.rs](seda-wasm-vm/libtallyvm/build.rs)
- [seda-wasm-vm/libtallyvm/cbindgen.toml](seda-wasm-vm/libtallyvm/cbindgen.toml)
- [seda-wasm-vm/libtallyvm/src/errors.rs](seda-wasm-vm/libtallyvm/src/errors.rs)
- [seda-wasm-vm/libtallyvm/src/lib.rs](seda-wasm-vm/libtallyvm/src/lib.rs)
- [seda-wasm-vm/runtime/core/src/context.rs](seda-wasm-vm/runtime/core/src/context.rs)
- [seda-wasm-vm/runtime/core/src/core_vm_imports/call_result.rs](seda-wasm-vm/runtime/core/src/core_vm_imports/call_result.rs)
- [seda-wasm-vm/runtime/core/src/core_vm_imports/execution_result.rs](seda-wasm-vm/runtime/core/src/core_vm_imports/execution_result.rs)
- [seda-wasm-vm/runtime/core/src/core_vm_imports/keccak256.rs](seda-wasm-vm/runtime/core/src/core_vm_imports/keccak256.rs)
- [seda-wasm-vm/runtime/core/src/core_vm_imports/mod.rs](seda-wasm-vm/runtime/core/src/core_vm_imports/mod.rs)
- [seda-wasm-vm/runtime/core/src/core_vm_imports/secp256_k1.rs](seda-wasm-vm/runtime/core/src/core_vm_imports/secp256_k1.rs)
- [seda-wasm-vm/runtime/core/src/errors.rs](seda-wasm-vm/runtime/core/src/errors.rs)
- [seda-wasm-vm/runtime/core/src/lib.rs](seda-wasm-vm/runtime/core/src/lib.rs)
- [seda-wasm-vm/runtime/core/src/metering.rs](seda-wasm-vm/runtime/core/src/metering.rs)
- [seda-wasm-vm/runtime/core/src/resources_dir.rs](seda-wasm-vm/runtime/core/src/resources_dir.rs)
- [seda-wasm-vm/runtime/core/src/runtime.rs](seda-wasm-vm/runtime/core/src/runtime.rs)
- [seda-wasm-vm/runtime/core/src/runtime_context.rs](seda-wasm-vm/runtime/core/src/runtime_context.rs)
- [seda-wasm-vm/runtime/core/src/safe_wasi_imports.rs](seda-wasm-vm/runtime/core/src/safe_wasi_imports.rs)
- [seda-wasm-vm/runtime/core/src/tally_vm_imports/http_fetch.rs](seda-wasm-vm/runtime/core/src/tally_vm_imports/http_fetch.rs)
- [seda-wasm-vm/runtime/core/src/tally_vm_imports/mod.rs](seda-wasm-vm/runtime/core/src/tally_vm_imports/mod.rs)
- [seda-wasm-vm/runtime/core/src/tally_vm_imports/proxy_http_fetch.rs](seda-wasm-vm/runtime/core/src/tally_vm_imports/proxy_http_fetch.rs)
- [seda-wasm-vm/runtime/core/src/vm_imports.rs](seda-wasm-vm/runtime/core/src/vm_imports.rs)
- [seda-wasm-vm/runtime/core/src/wasm_cache.rs](seda-wasm-vm/runtime/core/src/wasm_cache.rs)
- [seda-wasm-vm/tallyvm/execute.go](seda-wasm-vm/tallyvm/execute.go)
- [seda-wasm-vm/tallyvm/libseda_tally_vm.h](seda-wasm-vm/tallyvm/libseda_tally_vm.h)
- [seda-wasm-vm/tallyvm/link_glibclinux_aarch64.go](seda-wasm-vm/tallyvm/link_glibclinux_aarch64.go)
- [seda-wasm-vm/tallyvm/link_glibclinux_x86_64.go](seda-wasm-vm/tallyvm/link_glibclinux_x86_64.go)
- [seda-wasm-vm/tallyvm/link_mac.go](seda-wasm-vm/tallyvm/link_mac.go)
- [seda-wasm-vm/tallyvm/link_muslc_aarch64.go](seda-wasm-vm/tallyvm/link_muslc_aarch64.go)
- [seda-wasm-vm/tallyvm/link_muslc_x86_64.go](seda-wasm-vm/tallyvm/link_muslc_x86_64.go)


