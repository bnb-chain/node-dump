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
### 1. download from R2
mkdir -p ${NODE_DATA_PATH}
wget -qO- $NODE_DATA_LINK | tar -zxvf - -C ${NODE_DATA_PATH}

### 2. download from greenfield
#### Download the list of blockchain snapshot segment links.
wget $NODE_SEGMENT_LINKS -O ./bc-snapshot-segment-links.txt

#### Loop through each segment link in the list.
while read -r line; do
  #### Download the blockchain snapshot segment and append it to the main archive file.
  wget $line -O - >> bc-snapshot.tar.gz
done < ./bc-snapshot-segment-links.txt

#### Extract the blockchain snapshot to the specified data directory.
tar -xzvf bc-snapshot.tar.gz -C ${NODE_DATA_PATH}

## Merkle Proofs of User Accounts
mkdir -p ${ARCHIVED_PROOF_PATH}
wget -qO- $MERKLE_PROOF_DATA_LINK | tar -zxvf - -C ${ARCHIVED_PROOF_PATH}
```

## Verify Proofs Data

verify the merkle proofs data from the fullnode to ensure the merkle proofs is matching the state of the fullnode.

### Mainnet

```bash
./build/dump verify ${ARCHIVED_PROOF_PATH}/bc-mainnet-proofs --home $NODE_DATA_PATH/gaiad --tracelog
```

### Testnet

```bash
./build/dump verify ${ARCHIVED_PROOF_PATH}/dump --home $NODE_DATA_PATH/dataseed --tracelog
```
