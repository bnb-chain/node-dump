## BNB Beacon Chain Dump

## Introduction
Originally conceived as a platform for issuing data assets, the Beacon Chain (BC) has evolved to host 7.6 million accounts on the BNB Beacon Chain, supporting 557 tokens compliant with BEP2 or BEP8 standards. The digital assets held by users on this chain are secure and will persist beyond the BC Fusion event. The responsibility for safeguarding these assets falls upon the BNB Chain, irrespective of their individual values.

Our objective is to implement a solution that ensures the seamless execution of BC Fusion, followed by secure access to users' digital assets.

Following the BC-Fusion plan, the BNB Beacon Chain has been officially decommissioned.

This tool serves the purpose of dumping the state of the BNB Beacon Chain and generating Merkle tree proofs for user accounts.

## Archived Data

The following data is available for download:

### BNB Beacon Chain Node

#### Mainnet

| Field |Value |
| --- | --- |
| Chain ID | `Binance-Chain-Tigris` |
| Commit Hash | `JdLTQmMqSmhFQrdmX0/XvpyXWFvcrJ/9pXirC/RyDzk=` |
| Block | `385251927` |
| R2 Link | [Download](https://pub-c0627345c16f47ab858c9469133073a8.r2.dev/bc-mainnet-dataseed.tar.gz) |
| Greenfield Link | `TBD` |
| Size | 1.7T |
| SHA256 | `TBD` |

#### Testnet

| Field |Value |
| --- | --- |
| Chain ID | `Binance-Chain-Ganges` |
| Commit Hash | `LeswMibeF/8ao8md6hbmFYHVXg/E+zVxjKO376qLGXo=` |
| Block | `56503598` |
| R2 Link | [Download](https://pub-c0627345c16f47ab858c9469133073a8.r2.dev/bc-testnet-dataseed.tar.gz) |
| Greenfield Link | [Segment Download Links](https://raw.githubusercontent.com/bnb-chain/node-dump/blob/master/asset/bc-testnet-snapshot-segment-links.txt) |
| Size | 164G |
| SHA256 | `777a25f6d3228acb1854f1366b13befc1c2089ae2740cf5757120682ffc79a30` |

### Merkle Proofs of User Accounts

#### Mainnet

| Field |Value |
| --- | --- |
| Chain ID | `Binance-Chain-Tigris` |
| Commit Hash | `JdLTQmMqSmhFQrdmX0/XvpyXWFvcrJ/9pXirC/RyDzk=` |
| Block | `385251927` |
| R2 Link | [Download](https://pub-c0627345c16f47ab858c9469133073a8.r2.dev/bc-mainnet-proofs.tar.gz) |
| Greenfield Link | [Download](`https://greenfield-sp.nodereal.io/view/bnb-beacon-chain-archive/bc-mainnet-proofs.tar.gz`) |
| Size | 833M |
| SHA256 | `4fdf783b6cc5ba688775ed23f7e74651c95a2788b163a99e42770c356434e3e8` |

#### Testnet

| Field |Value |
| --- | --- |
| Chain ID | `Binance-Chain-Ganges` |
| Commit Hash | `LeswMibeF/8ao8md6hbmFYHVXg/E+zVxjKO376qLGXo=` |
| Block | `56503598` |
| R2 Link | [Download](https://pub-c0627345c16f47ab858c9469133073a8.r2.dev/bc-testnet-proofs.tar.gz) |
| Greenfield Link | [Download](https://greenfield-sp.nodereal.io/view/bnb-beacon-chain-archive/bc-testnet-proofs.tar.gz) |
| Size | 15M |
| SHA256 | `69cc59903e514c529018fafbdebba0bafc6f8e1ef8a2602d4ce573a314b2eb9a` |

## Verification

Please refer to the [verification guide](./docs/verification.md) for more details.
