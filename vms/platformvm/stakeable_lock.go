package platformvm

import (
	"errors"

	"github.com/djt-labs/paaro/vms/components/djtx"
)

var errInvalidLocktime = errors.New("invalid locktime")

type StakeableLockOut struct {
	Locktime             uint64 `serialize:"true" json:"locktime"`
	djtx.TransferableOut `serialize:"true"`
}

func (s *StakeableLockOut) Addresses() [][]byte {
	if addressable, ok := s.TransferableOut.(djtx.Addressable); ok {
		return addressable.Addresses()
	}
	return nil
}

func (s *StakeableLockOut) Verify() error {
	if s.Locktime == 0 {
		return errInvalidLocktime
	}
	if _, nested := s.TransferableOut.(*StakeableLockOut); nested {
		return errors.New("shouldn't nest stakeable locks")
	}
	return s.TransferableOut.Verify()
}

type StakeableLockIn struct {
	Locktime            uint64 `serialize:"true" json:"locktime"`
	djtx.TransferableIn `serialize:"true"`
}

func (s *StakeableLockIn) Verify() error {
	if s.Locktime == 0 {
		return errInvalidLocktime
	}
	if _, nested := s.TransferableIn.(*StakeableLockIn); nested {
		return errors.New("shouldn't nest stakeable locks")
	}
	return s.TransferableIn.Verify()
}
