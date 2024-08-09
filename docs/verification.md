# Verify Merkle Tree Proofs

This chapter describes how to verify the merkle tree proofs data from the fullnode to ensure the merkle proofs is matching the state of the fullnode.

## Prepare Verification Tool

```bash
git clone https://github.com/bnb-chain/node-dump.git
cd node-dump
make build
```

## Download the Archived Data

Get the link from the data provider and download the data to the local machine.
refer to [readme.md](../Readme.md) for more details.
- **NODE_DATA_PATH**: the path of the BNB Beacon Chain Node data
- **ARCHIVED_PROOF_PATH**: the path of the Merkle Proofs of User Accounts data

```bash
## BNB Beacon Chain Node
mkdir -p ${NODE_DATA_PATH}
wget $NODE_DATA_LINK -O - | tar -xzvf -C ${NODE_DATA_PATH}
## Merkle Proofs of User Accounts
mkdir -p ${ARCHIVED_PROOF_PATH}
wget $MERKLE_PROOF_DATA_LINK -O - | tar -xzvf -C ${ARCHIVED_PROOF_PATH}
```

## Verify Proofs Data

verify the merkle proofs data from the fullnode to ensure the merkle proofs is matching the state of the fullnode.
```bash
./build/dump verify ${ARCHIVED_PROOF_PATH}/dump --home $NODE_DATA_PATH/dataseed --tracelog
```
