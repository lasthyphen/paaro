// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gsharedmemory

import (
	"context"
	"io"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/lasthyphen/paaro/chains/atomic"
	"github.com/lasthyphen/paaro/chains/atomic/gsharedmemory/gsharedmemoryproto"
	"github.com/lasthyphen/paaro/database"
	"github.com/lasthyphen/paaro/database/memdb"
	"github.com/lasthyphen/paaro/database/prefixdb"
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/utils/logging"
	"github.com/lasthyphen/paaro/utils/units"
)

const (
	bufSize = units.MiB
)

func TestInterface(t *testing.T) {
	assert := assert.New(t)

	chainID0 := ids.GenerateTestID()
	chainID1 := ids.GenerateTestID()

	for _, test := range atomic.SharedMemoryTests {
		m := atomic.Memory{}
		baseDB := memdb.New()
		memoryDB := prefixdb.New([]byte{0}, baseDB)
		testDB := prefixdb.New([]byte{1}, baseDB)

		err := m.Initialize(logging.NoLog{}, memoryDB)
		assert.NoError(err)

		sm0, conn0 := wrapSharedMemory(t, m.NewSharedMemory(chainID0), baseDB)
		sm1, conn1 := wrapSharedMemory(t, m.NewSharedMemory(chainID1), baseDB)

		test(t, chainID0, chainID1, sm0, sm1, testDB)

		err = conn0.Close()
		assert.NoError(err)

		err = conn1.Close()
		assert.NoError(err)
	}
}

func wrapSharedMemory(t *testing.T, sm atomic.SharedMemory, db database.Database) (atomic.SharedMemory, io.Closer) {
	listener := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	gsharedmemoryproto.RegisterSharedMemoryServer(server, NewServer(sm, db))
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	dialer := grpc.WithContextDialer(
		func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		},
	)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", dialer, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial: %s", err)
	}

	rpcsm := NewClient(gsharedmemoryproto.NewSharedMemoryClient(conn))
	return rpcsm, conn
}
