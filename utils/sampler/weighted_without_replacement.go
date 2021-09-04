// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sampler

// WeightedWithoutReplacement defines how to sample weight without replacement.
// Note that the behavior is to sample the weight without replacement, not the
// indices. So duplicate indices can be returned.
type WeightedWithoutReplacement interface {
	Initialize(weights []uint64) error
	Sample(count int) ([]int, error)
}

// NewWeightedWithoutReplacement returns a new sampler
func NewWeightedWithoutReplacement() WeightedWithoutReplacement {
	return &weightedWithoutReplacementGeneric{
		u: NewUniform(),
		w: NewWeighted(),
	}
}

// NewBestWeightedWithoutReplacement returns a new sampler
func NewBestWeightedWithoutReplacement(
	expectedSampleSize int,
) WeightedWithoutReplacement {
	return &weightedWithoutReplacementGeneric{
		u: NewBestUniform(expectedSampleSize),
		w: NewWeighted(),
	}
}
