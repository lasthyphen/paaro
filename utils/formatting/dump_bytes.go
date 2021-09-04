// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package formatting

import (
	"encoding/hex"
	"strings"
)

type DumpBytes struct{ Bytes []byte }

func (db DumpBytes) String() string { return strings.TrimSpace(hex.Dump(db.Bytes)) }
