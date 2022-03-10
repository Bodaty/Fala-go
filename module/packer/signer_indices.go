package packer

import (
	"fmt"
)

// EncodeSignerIndices encodes indices into compacted bit vector. Each bit represents whether the identity at
// that index is a signer.
// It assumes the indices is in the increasing order,
// otherwise, decoding the signer indices will not recover to the original indices.
// TODO : check index in the indices are in the increasing order
func EncodeSignerIndices(indices []int, count int) []byte {
	totalBytes := bytesCount(count)
	bytes := make([]byte, totalBytes)
	for _, index := range indices {
		byt := index >> 3
		offset := 7 - (index & 7)
		mask := byte(1 << offset)
		bytes[byt] ^= mask
	}
	return bytes
}

// DecodeSignerIndices decodes the given compacted signer indices to a slice of indices.
func DecodeSignerIndices(indices []byte, count int) ([]int, error) {
	if bytesCount(count) != len(indices) {
		return nil, fmt.Errorf("signer indices has wrong count, expect count %v, but actually got %v",
			bytesCount(count), len(indices))
	}

	signerIndices := make([]int, 0, count)

	var byt byte
	var offset int

	for index := 0; index < count; index++ {
		byt = indices[index>>3]
		offset = 7 - (index & 7)
		mask := byte(1 << offset)
		if byt&mask > 0 {
			signerIndices = append(signerIndices, index)
		}
	}

	// remaining bits (if any), they must be all `0`s
	remainings := byt << (8 - offset)
	if remainings != byte(0) {
		return nil, fmt.Errorf("the remaining bites are expected to be all 0s, but are %v", remainings)
	}

	return signerIndices, nil
}

func bytesCount(count int) int {
	return (count + 7) >> 3
}
