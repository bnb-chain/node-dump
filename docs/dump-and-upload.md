# Dump Accounts and Generate Merkle Tree Proofs

This chapter describes how to dump the state of the BNB Beacon Chain and generate Merkle tree proofs for user accounts.

## Prepare The Tool and Download the Archived Data

```bash
## download the BNB Beacon Chain Node data
mkdir -p ${DATA_HOME}
wget $NODE_DATA_LINK -O - | tar -xz -C ${DATA_HOME}

## build the tool and dump the state to merkle proofs
make build
mkdir -p ./output
./build/dump export ./output/ --home ${DATA_HOME}

## tar and compress the output
tar -zcvf ArchivedMerkleProof.tar.gz ./output
```

## Upload to S3

```bash
aws s3 cp ./ArchivedMerkleProof.tar.gz s3://<bucket-name>/ArchivedMerkleProof.tar.gz
```

## Upload to Greenfield

```bash
gnfd-cmd object put ./ArchivedMerkleProof.tar.gz gnfd://<bucket-name>
```

## Upload to Arweave (coming soon)
