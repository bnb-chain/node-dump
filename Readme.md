## BNB Beacon Chain Dump

## Introduction
Originally conceived as a platform for issuing data assets, the Beacon Chain (BC) has evolved to host 7.6 million accounts on the BNB Beacon Chain, supporting 557 tokens compliant with BEP2 or BEP8 standards. The digital assets held by users on this chain are secure and will persist beyond the BC Fusion event. The responsibility for safeguarding these assets falls upon the BNB Chain, irrespective of their individual values.

Our objective is to implement a solution that ensures the seamless execution of BC Fusion, followed by secure access to users' digital assets.

Following the BC-Fusion plan, the BNB Beacon Chain has been officially decommissioned.

This tool serves the purpose of dumping the state of the BNB Beacon Chain and generating Merkle tree proofs for user accounts.

### Dump Accounts and Generate Merkle Tree Proofs
```bash
```bash
make build
mkdir -p ./output
./build/dump export ./output/state.json --home ${DATA_HOME}
```
