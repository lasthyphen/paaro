// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpcchainvm

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc"

	"github.com/hashicorp/go-plugin"

	"github.com/djt-labs/paaro/api/keystore/gkeystore"
	"github.com/djt-labs/paaro/api/keystore/gkeystore/gkeystoreproto"
	"github.com/djt-labs/paaro/chains/atomic/gsharedmemory"
	"github.com/djt-labs/paaro/chains/atomic/gsharedmemory/gsharedmemoryproto"
	"github.com/djt-labs/paaro/database/manager"
	"github.com/djt-labs/paaro/database/rpcdb"
	"github.com/djt-labs/paaro/database/rpcdb/rpcdbproto"
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/snow/choices"
	"github.com/djt-labs/paaro/snow/engine/common"
	"github.com/djt-labs/paaro/snow/engine/snowman/block"
	"github.com/djt-labs/paaro/utils/logging"
	"github.com/djt-labs/paaro/utils/wrappers"
	"github.com/djt-labs/paaro/version"
	"github.com/djt-labs/paaro/vms/rpcchainvm/galiaslookup"
	"github.com/djt-labs/paaro/vms/rpcchainvm/galiaslookup/galiaslookupproto"
	"github.com/djt-labs/paaro/vms/rpcchainvm/ghttp"
	"github.com/djt-labs/paaro/vms/rpcchainvm/ghttp/ghttpproto"
	"github.com/djt-labs/paaro/vms/rpcchainvm/grpcutils"
	"github.com/djt-labs/paaro/vms/rpcchainvm/gsubnetlookup"
	"github.com/djt-labs/paaro/vms/rpcchainvm/gsubnetlookup/gsubnetlookupproto"
	"github.com/djt-labs/paaro/vms/rpcchainvm/messenger"
	"github.com/djt-labs/paaro/vms/rpcchainvm/messenger/messengerproto"
	"github.com/djt-labs/paaro/vms/rpcchainvm/vmproto"
)

var _ vmproto.VMServer = &VMServer{}

// VMServer is a VM that is managed over RPC.
type VMServer struct {
	vmproto.UnimplementedVMServer
	vm     block.ChainVM
	broker *plugin.GRPCBroker

	serverCloser grpcutils.ServerCloser
	connCloser   wrappers.Closer

	ctx      *snow.Context
	toEngine chan common.Message
}

// NewServer returns a vm instance connected to a remote vm instance
func NewServer(vm block.ChainVM, broker *plugin.GRPCBroker) *VMServer {
	return &VMServer{
		vm:     vm,
		broker: broker,
	}
}

func (vm *VMServer) Initialize(_ context.Context, req *vmproto.InitializeRequest) (*vmproto.InitializeResponse, error) {
	subnetID, err := ids.ToID(req.SubnetID)
	if err != nil {
		return nil, err
	}
	chainID, err := ids.ToID(req.ChainID)
	if err != nil {
		return nil, err
	}
	nodeID, err := ids.ToShortID(req.NodeID)
	if err != nil {
		return nil, err
	}
	xChainID, err := ids.ToID(req.XChainID)
	if err != nil {
		return nil, err
	}
	djtxAssetID, err := ids.ToID(req.DjtxAssetID)
	if err != nil {
		return nil, err
	}

	epochFirstTransition := time.Time{}
	if err := epochFirstTransition.UnmarshalBinary(req.EpochFirstTransition); err != nil {
		return nil, err
	}

	// Dial each database in the request and construct the database manager
	versionedDBs := make([]*manager.VersionedDatabase, len(req.DbServers))
	versionParser := version.NewDefaultParser()
	for i, vDBReq := range req.DbServers {
		version, err := versionParser.Parse(vDBReq.Version)
		if err != nil {
			// Ignore closing errors to return the original error
			_ = vm.connCloser.Close()
			return nil, err
		}

		dbConn, err := vm.broker.Dial(vDBReq.DbServer)
		if err != nil {
			// Ignore closing errors to return the original error
			_ = vm.connCloser.Close()
			return nil, err
		}
		vm.connCloser.Add(dbConn)

		versionedDBs[i] = &manager.VersionedDatabase{
			Database: rpcdb.NewClient(rpcdbproto.NewDatabaseClient(dbConn)),
			Version:  version,
		}
	}
	dbManager, err := manager.NewManagerFromDBs(versionedDBs)
	if err != nil {
		// Ignore closing errors to return the original error
		_ = vm.connCloser.Close()
		return nil, err
	}

	msgConn, err := vm.broker.Dial(req.EngineServer)
	if err != nil {
		// Ignore closing errors to return the original error
		_ = vm.connCloser.Close()
		return nil, err
	}
	vm.connCloser.Add(msgConn)

	keystoreConn, err := vm.broker.Dial(req.KeystoreServer)
	if err != nil {
		// Ignore closing error to return the original error
		_ = vm.connCloser.Close()
		return nil, err
	}
	vm.connCloser.Add(keystoreConn)

	sharedMemoryConn, err := vm.broker.Dial(req.SharedMemoryServer)
	if err != nil {
		// Ignore closing error to return the original error
		_ = vm.connCloser.Close()
		return nil, err
	}
	vm.connCloser.Add(sharedMemoryConn)

	bcLookupConn, err := vm.broker.Dial(req.BcLookupServer)
	if err != nil {
		// Ignore closing error to return the original error
		_ = vm.connCloser.Close()
		return nil, err
	}
	vm.connCloser.Add(bcLookupConn)

	snLookupConn, err := vm.broker.Dial(req.SnLookupServer)
	if err != nil {
		// Ignore closing error to return the original error
		_ = vm.connCloser.Close()
		return nil, err
	}

	msgClient := messenger.NewClient(messengerproto.NewMessengerClient(msgConn))
	keystoreClient := gkeystore.NewClient(gkeystoreproto.NewKeystoreClient(keystoreConn), vm.broker)
	sharedMemoryClient := gsharedmemory.NewClient(gsharedmemoryproto.NewSharedMemoryClient(sharedMemoryConn))
	bcLookupClient := galiaslookup.NewClient(galiaslookupproto.NewAliasLookupClient(bcLookupConn))
	snLookupClient := gsubnetlookup.NewClient(gsubnetlookupproto.NewSubnetLookupClient(snLookupConn))

	toEngine := make(chan common.Message, 1)
	go func() {
		for msg := range toEngine {
			// Nothing to do with the error within the goroutine
			_ = msgClient.Notify(msg)
		}
	}()

	vm.ctx = &snow.Context{
		NetworkID:            req.NetworkID,
		SubnetID:             subnetID,
		ChainID:              chainID,
		NodeID:               nodeID,
		XChainID:             xChainID,
		DJTXAssetID:          djtxAssetID,
		Log:                  logging.NoLog{},
		DecisionDispatcher:   nil,
		ConsensusDispatcher:  nil,
		Keystore:             keystoreClient,
		SharedMemory:         sharedMemoryClient,
		BCLookup:             bcLookupClient,
		SNLookup:             snLookupClient,
		EpochFirstTransition: epochFirstTransition,
		EpochDuration:        time.Duration(req.EpochDuration),
	}

	if err := vm.vm.Initialize(vm.ctx, dbManager, req.GenesisBytes, req.UpgradeBytes, req.ConfigBytes, toEngine, nil); err != nil {
		// Ignore errors closing resources to return the original error
		_ = vm.connCloser.Close()
		close(toEngine)
		return nil, err
	}

	vm.toEngine = toEngine
	lastAccepted, err := vm.vm.LastAccepted()
	if err != nil {
		return nil, err
	}
	blk, err := vm.vm.GetBlock(lastAccepted)
	if err != nil {
		return nil, err
	}
	parentID := blk.Parent()
	return &vmproto.InitializeResponse{
		LastAcceptedID:       lastAccepted[:],
		LastAcceptedParentID: parentID[:],
		Status:               uint32(choices.Accepted),
		Height:               blk.Height(),
		Bytes:                blk.Bytes(),
	}, err
}

func (vm *VMServer) Bootstrapping(context.Context, *vmproto.BootstrappingRequest) (*vmproto.BootstrappingResponse, error) {
	return &vmproto.BootstrappingResponse{}, vm.vm.Bootstrapping()
}

func (vm *VMServer) Bootstrapped(context.Context, *vmproto.BootstrappedRequest) (*vmproto.BootstrappedResponse, error) {
	vm.ctx.Bootstrapped()
	return &vmproto.BootstrappedResponse{}, vm.vm.Bootstrapped()
}

func (vm *VMServer) Shutdown(context.Context, *vmproto.ShutdownRequest) (*vmproto.ShutdownResponse, error) {
	if vm.toEngine == nil {
		return &vmproto.ShutdownResponse{}, nil
	}

	errs := wrappers.Errs{}
	errs.Add(vm.vm.Shutdown())
	close(vm.toEngine)

	vm.serverCloser.Stop()
	errs.Add(vm.connCloser.Close())

	return &vmproto.ShutdownResponse{}, errs.Err
}

func (vm *VMServer) CreateStaticHandlers(context.Context, *vmproto.CreateStaticHandlersRequest) (*vmproto.CreateStaticHandlersResponse, error) {
	handlers, err := vm.vm.CreateStaticHandlers()
	if err != nil {
		return nil, err
	}
	resp := &vmproto.CreateStaticHandlersResponse{}
	for prefix, h := range handlers {
		handler := h

		// start the messenger server
		serverID := vm.broker.NextId()
		go vm.broker.AcceptAndServe(serverID, func(opts []grpc.ServerOption) *grpc.Server {
			server := grpc.NewServer(opts...)
			vm.serverCloser.Add(server)
			ghttpproto.RegisterHTTPServer(server, ghttp.NewServer(handler.Handler, vm.broker))
			return server
		})

		resp.Handlers = append(resp.Handlers, &vmproto.Handler{
			Prefix:      prefix,
			LockOptions: uint32(handler.LockOptions),
			Server:      serverID,
		})
	}
	return resp, nil
}

func (vm *VMServer) CreateHandlers(_ context.Context, req *vmproto.CreateHandlersRequest) (*vmproto.CreateHandlersResponse, error) {
	handlers, err := vm.vm.CreateHandlers()
	if err != nil {
		return nil, err
	}
	resp := &vmproto.CreateHandlersResponse{}
	for prefix, h := range handlers {
		handler := h

		// start the messenger server
		serverID := vm.broker.NextId()
		go vm.broker.AcceptAndServe(serverID, func(opts []grpc.ServerOption) *grpc.Server {
			server := grpc.NewServer(opts...)
			vm.serverCloser.Add(server)
			ghttpproto.RegisterHTTPServer(server, ghttp.NewServer(handler.Handler, vm.broker))
			return server
		})

		resp.Handlers = append(resp.Handlers, &vmproto.Handler{
			Prefix:      prefix,
			LockOptions: uint32(handler.LockOptions),
			Server:      serverID,
		})
	}
	return resp, nil
}

func (vm *VMServer) BuildBlock(_ context.Context, _ *vmproto.BuildBlockRequest) (*vmproto.BuildBlockResponse, error) {
	blk, err := vm.vm.BuildBlock()
	if err != nil {
		return nil, err
	}
	blkID := blk.ID()
	parentID := blk.Parent()
	return &vmproto.BuildBlockResponse{
		Id:       blkID[:],
		ParentID: parentID[:],
		Bytes:    blk.Bytes(),
		Height:   blk.Height(),
	}, nil
}

func (vm *VMServer) ParseBlock(_ context.Context, req *vmproto.ParseBlockRequest) (*vmproto.ParseBlockResponse, error) {
	blk, err := vm.vm.ParseBlock(req.Bytes)
	if err != nil {
		return nil, err
	}
	blkID := blk.ID()
	parentID := blk.Parent()
	return &vmproto.ParseBlockResponse{
		Id:       blkID[:],
		ParentID: parentID[:],
		Status:   uint32(blk.Status()),
		Height:   blk.Height(),
	}, nil
}

func (vm *VMServer) GetBlock(_ context.Context, req *vmproto.GetBlockRequest) (*vmproto.GetBlockResponse, error) {
	id, err := ids.ToID(req.Id)
	if err != nil {
		return nil, err
	}
	blk, err := vm.vm.GetBlock(id)
	if err != nil {
		return nil, err
	}
	parentID := blk.Parent()
	return &vmproto.GetBlockResponse{
		ParentID: parentID[:],
		Bytes:    blk.Bytes(),
		Status:   uint32(blk.Status()),
		Height:   blk.Height(),
	}, nil
}

func (vm *VMServer) SetPreference(_ context.Context, req *vmproto.SetPreferenceRequest) (*vmproto.SetPreferenceResponse, error) {
	id, err := ids.ToID(req.Id)
	if err != nil {
		return nil, err
	}
	return &vmproto.SetPreferenceResponse{}, vm.vm.SetPreference(id)
}

func (vm *VMServer) Health(_ context.Context, req *vmproto.HealthRequest) (*vmproto.HealthResponse, error) {
	details, err := vm.vm.HealthCheck()
	if err != nil {
		return &vmproto.HealthResponse{}, err
	}

	// Try to stringify the details
	detailsStr := "couldn't parse health check details to string"
	switch details := details.(type) {
	case nil:
		detailsStr = ""
	case string:
		detailsStr = details
	case map[string]string:
		asJSON, err := json.Marshal(details)
		if err != nil {
			detailsStr = string(asJSON)
		}
	case []byte:
		detailsStr = string(details)
	}

	return &vmproto.HealthResponse{
		Details: detailsStr,
	}, nil
}

func (vm *VMServer) Version(_ context.Context, req *vmproto.VersionRequest) (*vmproto.VersionResponse, error) {
	version, err := vm.vm.Version()
	return &vmproto.VersionResponse{
		Version: version,
	}, err
}

func (vm *VMServer) BlockVerify(_ context.Context, req *vmproto.BlockVerifyRequest) (*vmproto.BlockVerifyResponse, error) {
	blk, err := vm.vm.ParseBlock(req.Bytes)
	if err != nil {
		return nil, err
	}
	return &vmproto.BlockVerifyResponse{}, blk.Verify()
}

func (vm *VMServer) BlockAccept(_ context.Context, req *vmproto.BlockAcceptRequest) (*vmproto.BlockAcceptResponse, error) {
	id, err := ids.ToID(req.Id)
	if err != nil {
		return nil, err
	}
	blk, err := vm.vm.GetBlock(id)
	if err != nil {
		return nil, err
	}
	if err := blk.Accept(); err != nil {
		return nil, err
	}
	return &vmproto.BlockAcceptResponse{}, nil
}

func (vm *VMServer) BlockReject(_ context.Context, req *vmproto.BlockRejectRequest) (*vmproto.BlockRejectResponse, error) {
	id, err := ids.ToID(req.Id)
	if err != nil {
		return nil, err
	}
	blk, err := vm.vm.GetBlock(id)
	if err != nil {
		return nil, err
	}
	if err := blk.Reject(); err != nil {
		return nil, err
	}
	return &vmproto.BlockRejectResponse{}, nil
}
