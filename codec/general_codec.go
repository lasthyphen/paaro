// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package codec

// GeneralCodec marshals and unmarshals structs including interfaces
type GeneralCodec interface {
	Codec
	Registry
}
