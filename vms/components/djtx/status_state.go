// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package djtx

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/lasthyphen/paaro/cache"
	"github.com/lasthyphen/paaro/cache/metercacher"
	"github.com/lasthyphen/paaro/database"
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow/choices"
)

const (
	statusCacheSize = 8192
)

// StatusState is a thin wrapper around a database to provide, caching,
// serialization, and de-serialization for statuses.
type StatusState interface {
	// Status returns a status from storage.
	GetStatus(id ids.ID) (choices.Status, error)

	// PutStatus saves a status in storage.
	PutStatus(id ids.ID, status choices.Status) error

	// DeleteStatus removes a status from storage.
	DeleteStatus(id ids.ID) error
}

type statusState struct {
	// ID -> Status of thing with that ID, or nil if StatusState doesn't have
	// that status.
	statusCache cache.Cacher
	statusDB    database.Database
}

func NewStatusState(db database.Database) StatusState {
	return &statusState{
		statusCache: &cache.LRU{Size: statusCacheSize},
		statusDB:    db,
	}
}

func NewMeteredStatusState(db database.Database, namespace string, metrics prometheus.Registerer) (StatusState, error) {
	cache, err := metercacher.New(
		fmt.Sprintf("%s_status_cache", namespace),
		metrics,
		&cache.LRU{Size: statusCacheSize},
	)
	return &statusState{
		statusCache: cache,
		statusDB:    db,
	}, err
}

func (s *statusState) GetStatus(id ids.ID) (choices.Status, error) {
	if statusIntf, found := s.statusCache.Get(id); found {
		if statusIntf == nil {
			return choices.Unknown, database.ErrNotFound
		}
		return statusIntf.(choices.Status), nil
	}

	val, err := database.GetUInt32(s.statusDB, id[:])
	if err == database.ErrNotFound {
		s.statusCache.Put(id, nil)
		return choices.Unknown, database.ErrNotFound
	}
	if err != nil {
		return choices.Unknown, err
	}

	status := choices.Status(val)
	if err := status.Valid(); err != nil {
		return choices.Unknown, err
	}

	s.statusCache.Put(id, status)
	return status, nil
}

func (s *statusState) PutStatus(id ids.ID, status choices.Status) error {
	s.statusCache.Put(id, status)
	return database.PutUInt32(s.statusDB, id[:], uint32(status))
}

func (s *statusState) DeleteStatus(id ids.ID) error {
	s.statusCache.Put(id, nil)
	return s.statusDB.Delete(id[:])
}
