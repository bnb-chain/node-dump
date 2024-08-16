package util

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestVerifyMerkleProof(t *testing.T) {
	root := MustDecodeHexToBytes("0x59bb94f7047904a8fdaec42e4785295167f7fd63742b309afeb84bd71f8e6554")
	proof := MustDecodeHexArrayToBytes([]string{
		"0x061680518f3f97c075a62df766fa55c90b0c415140f737c0d1f7ace5ad2bfee6",
		"0x366f06cef0f1668d848819cb7b5a07b0093ad997da496e60060db2fee754857b",
		"0x6debec5a4272951843cf24f74c30d5ccf1afec9aafbfc45d0b50cb4eb6f89c09",
		"0x5cb2e4d880e2387764df4de9ce49cbabc41b6e4a07b1c2e1d9fc98957b6643d2",
		"0x88c6195b4444035bef3212847f38822c0d509d811de8c9154e7f5f8ec3778b67",
		"0x27c985cced25522043ded2fc8103baa24edc21b6c9f95c5bfff635ab36bdb29d",
		"0x39a0fbfba925ebd0cf4f5fe5ab4c69eb18317fd1bd4373647a53dc339fb764a9",
		"0x61300a7a7fe0932760c1e1edfa4d4450cc378d9b5c538dcb24ffbbc18f249fe5",
		"0x4d49fcf8a1e0b72b535921dea8e02baac18df614e7f7c462749a2b14ee2737ef",
		"0xc10261d3337346f921c4fef13ba1bcb46a531e947ce41c81e54404e970deaaf5",
		"0x3536a24678835b0f7adeae1f27dae7d6bb22598fb8f8578ec0eef5ea5146f85b",
		"0x925aab793d8080c4f8ea5034e195938c5550f7ba80acf7d7e7d8468f5b5dd70a",
		"0xdef2b6210654ac4f48b4556e24907e027e66729045d0c669a53c75a880477b48",
		"0x4bb1aab890245e6a9e1e969ae3f6f0315ea073606fd6fabe9f3d7514c84fee98",
		"0xe096d4b3669b1c7cd8fcff26b2b00029c09c0f38a34ae632b022622fb46ad69a",
		"0x05e63b558cba63f5add60201151f96ff8f5370d2b8280a96b4fa8fd2d519ab9f",
		"0xa2d456e52facaa953bfbc79a5a6ed7647dda59872b9b35c20183887eeb4640eb"})
	leaf := MustDecodeHexToBytes("0xe7b660e08a0bf3b78615c3a9d6804c31d6e29371e6dcde4280e5484ac8d18c86")

	if !VerifyMerkleProof(root, proof, leaf) {
		t.Error("VerifyMerkleProof failed")
		t.Logf("root: %s", hexutil.Encode(root))
		hash := leaf
		for _, proofElement := range proof {
			hash = hashPair(hash, proofElement)
		}
		t.Logf("hash: %s", hexutil.Encode(hash))
	} else {
		t.Log("VerifyMerkleProof - passed")
	}

	// wrong leaf bytes
	proof = MustDecodeHexArrayToBytes([]string{
		"0x061680518f3f97c075a62df766fa55c90b0c415140f737c0d1f7ace5ad2bfee6",
		"0x366f06cef0f1668d848819cb7b5a07b0093ad997da496e60060db2fee754857b",
		"0x6debec5a4272951843cf24f74c30d5ccf1afec9aafbfc45d0b50cb4eb6f89c09",
		"0x5cb2e4d880e2387764df4de9ce49cbabc41b6e4a07b1c2e1d9fc98957b6643d2",
		"0x88c6195b4444035bef3212847f38822c0d509d811de8c9154e7f5f8ec3778b67",
		"0x27c985cced25522043ded2fc8103baa24edc21b6c9f95c5bfff635ab36bdb29d",
		"0x39a0fbfba925ebd0cf4f5fe5ab4c69eb18317fd1bd4373647a53dc339fb764a9",
		"0x61300a7a7fe0932760c1e1edfa4d4450cc378d9b5c538dcb24ffbbc18f249fe5",
		"0x4d49fcf8a1e0b72b535921dea8e02baac18df614e7f7c462749a2b14ee2737ef",
		"0xc10261d3337346f921c4fef13ba1bcb46a531e947ce41c81e54404e970deaaf5",
		"0x3536a24678835b0f7adeae1f27dae7d6bb22598fb8f8578ec0eef5ea5146f85b",
		"0x925aab793d8080c4f8ea5034e195938c5550f7ba80acf7d7e7d8468f5b5dd70a",
		"0xdef2b6210654ac4f48b4556e24907e027e66729045d0c669a53c75a880477b48",
		"0x4bb1aab890245e6a9e1e969ae3f6f0315ea073606fd6fabe9f3d7514c84fee98",
		"0xe096d4b3669b1c7cd8fcff26b2b00029c09c0f38a34ae632b022622fb46ad69a",
		"0x05e63b558cba63f5add60201151f96ff8f5370d2b8280a96b4fa8fd2d519ab9f",
		"0xa2d456e52facaa953bfbc79a5a6ed7647dda59872b9b35c20183887eeb4640eb"})
	leaf = MustDecodeHexToBytes("0x111160e08a0bf3b78615c3a9d6804c31d6e29371e6dcde4280e5484ac8d18c86")

	if VerifyMerkleProof(root, proof, leaf) {
		t.Error("VerifyMerkleProof should fail")
		t.Logf("root: %s", hexutil.Encode(root))
		hash := leaf
		for _, proofElement := range proof {
			hash = hashPair(hash, proofElement)
		}
		t.Logf("hash: %s", hexutil.Encode(hash))
	} else {
		t.Log("VerifyMerkleProof - wrong leaf bytes passed")
	}

	// wrong proof bytes
	proof = MustDecodeHexArrayToBytes([]string{
		"0x061680518f3f97c075a62df766fa55c90b0c415140f737c0d1f7ace5ad2bfee6",
		"0x366f06cef0f1668d848819cb7b5a07b0093ad997da496e60060db2fee754857b",
		"0x6debec5a4272951843cf24f74c30d5ccf1afec9aafbfc45d0b50cb4eb6f89c09",
		"0x5cb2e4d880e2387764df4de9ce49cbabc41b6e4a07b1c2e1d9fc98957b6643d2",
		"0x88c6195b4444035bef3212847f38822c0d509d811de8c9154e7f5f8ec3778b67",
		"0x27c985cced25522043ded2fc8103baa24edc21b6c9f95c5bfff635ab36bdb29d",
		"0x39a0fbfba925ebd0cf4f5fe5ab4c69eb18317fd1bd4373647a53dc339fb764a9",
		"0x12344a7a7fe0932760c1e1edfa4d4450cc378d9b5c538dcb24ffbbc18f249fe5",
		"0x4d49fcf8a1e0b72b535921dea8e02baac18df614e7f7c462749a2b14ee2737ef",
		"0xc10261d3337346f921c4fef13ba1bcb46a531e947ce41c81e54404e970deaaf5",
		"0x3536a24678835b0f7adeae1f27dae7d6bb22598fb8f8578ec0eef5ea5146f85b",
		"0x925aab793d8080c4f8ea5034e195938c5550f7ba80acf7d7e7d8468f5b5dd70a",
		"0xdef2b6210654ac4f48b4556e24907e027e66729045d0c669a53c75a880477b48",
		"0x4bb1aab890245e6a9e1e969ae3f6f0315ea073606fd6fabe9f3d7514c84fee98",
		"0xe096d4b3669b1c7cd8fcff26b2b00029c09c0f38a34ae632b022622fb46ad69a",
		"0x05e63b558cba63f5add60201151f96ff8f5370d2b8280a96b4fa8fd2d519ab9f",
		"0xa2d456e52facaa953bfbc79a5a6ed7647dda59872b9b35c20183887eeb4640eb"})
	leaf = MustDecodeHexToBytes("0xe7b660e08a0bf3b78615c3a9d6804c31d6e29371e6dcde4280e5484ac8d18c86")

	if VerifyMerkleProof(root, proof, leaf) {
		t.Error("VerifyMerkleProof should fail")
		t.Logf("root: %s", hexutil.Encode(root))
		hash := leaf
		for _, proofElement := range proof {
			hash = hashPair(hash, proofElement)
		}
		t.Logf("hash: %s", hexutil.Encode(hash))
	} else {
		t.Log("VerifyMerkleProof - wrong proof bytes passed")
	}
}
