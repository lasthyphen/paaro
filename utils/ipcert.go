// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import "crypto/x509"

type IPCertDesc struct {
	Cert      *x509.Certificate
	IPDesc    IPDesc
	Time      uint64
	Signature []byte
}
