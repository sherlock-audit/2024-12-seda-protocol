use std::collections::HashMap;

use cosmwasm_std::{coins, from_json, testing::MockApi, to_json_binary, Addr, StdError};
use cw_multi_test::{App, AppBuilder, ContractWrapper, Executor};
use k256::{
    ecdsa::{SigningKey, VerifyingKey},
    elliptic_curve::rand_core::OsRng,
};
use seda_common::{msgs::*, types::ToHexStr};
use serde::{de::DeserializeOwned, Serialize};
use sha3::{Digest, Keccak256};
use vrf_rs::Secp256k1Sha256;

use crate::{common_types::Hash, contract::*, error::ContractError, types::PublicKey};

pub fn new_public_key() -> (SigningKey, PublicKey) {
    let signing_key = SigningKey::random(&mut OsRng);
    let verifying_key = VerifyingKey::from(&signing_key);
    let public_key = verifying_key.to_encoded_point(true).as_bytes().try_into().unwrap();

    (signing_key, PublicKey(public_key))
}

pub struct TestInfo {
    app:           App,
    contract_addr: Addr,
    executors:     HashMap<&'static str, TestExecutor>,
    chain_id:      String,
}

impl TestInfo {
    pub fn init() -> Self {
        let mut executors = HashMap::new();
        let mut app = AppBuilder::default()
            .with_api(MockApi::default().with_prefix("seda"))
            .build(|router, api, storage| {
                let creator_addr = api.addr_make("creator");
                let creator = TestExecutor::new("creator", creator_addr.clone());
                router
                    .bank
                    .init_balance(storage, &creator_addr, coins(1_000_000, "aseda"))
                    .unwrap();
                executors.insert("creator", creator);
            });
        let contract = Box::new(ContractWrapper::new(execute, instantiate, query).with_sudo(sudo));
        let chain_id = "seda_test".to_string();
        let creator = executors.get("creator").unwrap();

        let code_id = app.store_code_with_creator(creator.addr(), contract);
        let init_msg = &InstantiateMsg {
            token:          "aseda".to_string(),
            owner:          creator.addr().into_string(),
            chain_id:       chain_id.clone(),
            staking_config: None,
            timeout_config: None,
        };

        let contract_addr = app
            .instantiate_contract(code_id, creator.addr(), &init_msg, &[], "core", None)
            .unwrap();

        let mut info = Self {
            app,
            contract_addr,
            executors,
            chain_id,
        };

        info.set_block_height(0);

        info
    }

    pub fn new_address(&mut self, name: &'static str) -> Addr {
        self.app.api().addr_make(name)
    }

    pub fn new_executor(&mut self, name: &'static str, amount: Option<u128>) -> TestExecutor {
        let addr = self.new_address(name);
        let executor = TestExecutor::new(name, addr);
        self.executors.insert(name, executor);
        let executor = self.executor(name).clone();

        if let Some(amount) = amount {
            self.app.init_modules(|router, _api, storage| {
                router
                    .bank
                    .init_balance(storage, &executor.addr, coins(amount, "aseda"))
                    .unwrap();
            });
        }
        executor
    }

    #[track_caller]
    pub fn executor(&self, name: &'static str) -> &TestExecutor {
        self.executors.get(name).unwrap()
    }

    #[track_caller]
    pub fn executor_balance(&self, name: &'static str) -> u128 {
        let executor = self.executors.get(name).unwrap();
        self.app()
            .wrap()
            .query_balance(executor.addr(), "aseda")
            .unwrap()
            .amount
            .u128()
    }

    pub fn app(&self) -> &App {
        &self.app
    }

    pub fn app_mut(&mut self) -> &mut App {
        &mut self.app
    }

    pub fn chain_id(&self) -> &str {
        self.chain_id.as_str()
    }

    pub fn block_height(&mut self) -> u64 {
        self.app.block_info().height
    }

    pub fn set_block_height(&mut self, height: u64) {
        self.app.update_block(|b| {
            b.height = height;
        });
    }

    pub fn creator(&self) -> TestExecutor {
        self.executor("creator").clone()
    }

    pub fn contract_addr(&self) -> Addr {
        self.contract_addr.clone()
    }

    pub fn contract_addr_str(&self) -> &str {
        self.contract_addr.as_str()
    }

    pub fn contract_addr_bytes(&self) -> &[u8] {
        self.contract_addr.as_bytes()
    }

    pub fn query<M: Serialize, R: DeserializeOwned>(&self, msg: M) -> Result<R, cosmwasm_std::StdError> {
        self.app.wrap().query_wasm_smart(self.contract_addr_str(), &msg)
    }

    #[track_caller]
    pub fn sudo<R: DeserializeOwned>(&mut self, msg: &SudoMsg) -> Result<R, ContractError> {
        let res = self.app.wasm_sudo(self.contract_addr.clone(), msg).map_err(|e| {
            if e.downcast_ref::<ContractError>().is_some() {
                e.downcast().unwrap()
            } else if let Some(s_err) = e.downcast_ref::<StdError>() {
                return ContractError::Std(s_err.to_string());
            } else {
                ContractError::Dbg(e.to_string())
            }
        });

        Ok(match res?.data {
            Some(data) => from_json(data).unwrap(),
            None => from_json(to_json_binary(&serde_json::Value::Null).unwrap()).unwrap(),
        })
    }

    #[track_caller]
    pub fn execute<R: DeserializeOwned>(
        &mut self,
        sender: &TestExecutor,
        msg: &ExecuteMsg,
    ) -> Result<R, ContractError> {
        let res = self
            .app
            .execute_contract(sender.addr(), self.contract_addr.clone(), msg, &[])
            .map_err(|e| {
                if e.downcast_ref::<ContractError>().is_some() {
                    e.downcast().unwrap()
                } else if let Some(s_err) = e.downcast_ref::<StdError>() {
                    return ContractError::Std(s_err.to_string());
                } else {
                    ContractError::Dbg(e.to_string())
                }
            });

        Ok(match res?.data {
            Some(data) => from_json(data).unwrap(),
            None => from_json(to_json_binary(&serde_json::Value::Null).unwrap()).unwrap(),
        })
    }

    #[track_caller]
    pub fn execute_with_funds<R: DeserializeOwned>(
        &mut self,
        sender: &mut TestExecutor,
        msg: &ExecuteMsg,
        amount: u128,
    ) -> Result<R, ContractError> {
        let res = self
            .app
            .execute_contract(sender.addr(), self.contract_addr.clone(), msg, &coins(amount, "aseda"))
            .map_err(|e| {
                if e.downcast_ref::<ContractError>().is_some() {
                    e.downcast().unwrap()
                } else if let Some(s_err) = e.downcast_ref::<StdError>() {
                    return ContractError::Std(s_err.to_string());
                } else {
                    ContractError::Dbg(e.to_string())
                }
            });

        Ok(match res?.data {
            Some(data) => from_json(data).unwrap(),
            None => from_json(to_json_binary(&serde_json::Value::Null).unwrap()).unwrap(),
        })
    }
}

#[derive(Debug, Clone)]
pub struct TestExecutor {
    pub name:    &'static str,
    addr:        Addr,
    signing_key: SigningKey,
    public_key:  PublicKey,
}

impl TestExecutor {
    fn new(name: &'static str, addr: Addr) -> Self {
        let (signing_key, public_key) = new_public_key();

        TestExecutor {
            name,
            addr,
            signing_key,
            public_key,
        }
    }

    pub fn addr(&self) -> Addr {
        self.addr.clone()
    }

    pub fn pub_key(&self) -> PublicKey {
        self.public_key.clone()
    }

    pub fn pub_key_hex(&self) -> String {
        self.public_key.to_hex()
    }

    pub fn sign_key(&self) -> Vec<u8> {
        self.signing_key.to_bytes().to_vec()
    }

    pub fn prove(&self, hash: &[u8]) -> Vec<u8> {
        let vrf = Secp256k1Sha256::default();
        vrf.prove(&self.signing_key.to_bytes(), hash).unwrap()
    }

    pub fn prove_hex(&self, hash: &[u8]) -> String {
        self.prove(hash).to_hex()
    }

    pub fn salt(&self) -> String {
        let mut hasher = Keccak256::new();
        hasher.update(self.name);
        let hash: Hash = hasher.finalize().into();
        hash.to_hex()
    }

    pub fn stake(&mut self, test_info: &mut TestInfo, amount: u128) -> Result<(), ContractError> {
        test_info.stake(self, None, amount)?;
        Ok(())
    }

    pub fn unstake(&mut self, test_info: &mut TestInfo, amount: u128) -> Result<(), ContractError> {
        test_info.unstake(self, amount)?;
        Ok(())
    }
}

#[test]
fn instantiate_works() {
    TestInfo::init();
}
