# Verify Merkle Tree Proofs

This chapter describes how to verify the merkle tree proofs data from the fullnode to ensure the merkle proofs is matching the state of the fullnode.

## Prepare Verification Tool

```bash
git clone https://github.com/bnb-chain/token-recover-approver.git
cd token-recover-approver
make build
```

## Download the Archived Data

Get the link from the data provider and download the data to the local machine.
refer to [readme.md](../Readme.md) for more details.

```bash
## BNB Beacon Chain Node
mkdir -p ${NODE_DATA_PATH}
wget $NODE_DATA_LINK -O - | tar -xz -C ${NODE_DATA_PATH}
## Merkle Proofs of User Accounts
mkdir -p ${ARCHIVED_PROOF_PATH}
wget $MERKLE_PROOF_DATA_LINK -O - | tar -xz -C ${ARCHIVED_PROOF_PATH}
```

## Verify Proofs Data

verify the merkle proofs data from the fullnode to ensure the merkle proofs is matching the state of the fullnode.

- **CHAIN_ID**: the chain id of the BNB Beacon Chain(Mainnet: Binance-Chain-Tigris, Testnet: Binance-Chain-Ganges)
- **MERKLE_ROOT**: the merkle root of the merkle proofs(stored in ${ARCHIVED_PROOF_PATH}/base.json)
- **STORE_MEMORY_STORE_MERKLE_PROOFS**: the path of the merkle proofs data
- **NODE_DATA_PATH**: the path of the BNB Beacon Chain Node data

```bash
export CHAIN_ID=Binance-Chain-Ganges
export MERKLE_ROOT=$(cat ${ARCHIVED_PROOF_PATH}/base.json | jq -r '.state_root')
export STORE_MEMORY_STORE_MERKLE_PROOFS=${ARCHIVED_PROOF_PATH}/proofs.json
./token-recover-approver/build/bin/approver tool verify-data-from-fullnode --home ${NODE_DATA_PATH} --verify_merkle_root
```
