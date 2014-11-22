package internal

import (
	"github.com/101loops/hrd/internal/types"
	"github.com/qedus/nds"

	ae "appengine"
	ds "appengine/datastore"
)

var (
	ndsDel = func(ctx ae.Context, keys []*ds.Key) error {
		return nds.DeleteMulti(ctx, keys)
	}
)

// DSDelete deletes the given entity.
func DSDelete(kind *types.Kind, src interface{}, multi bool) error {
	var err error
	var keys []*types.Key

	if multi {
		keys, err = types.GetEntitiesKeys(kind, src)
	} else {
		var key *types.Key
		key, err = types.GetEntityKey(kind, src)
		keys = []*types.Key{key}
	}

	if err != nil {
		return err
	}

	return DSDeleteKeys(kind, keys)
}

// DSDeleteKeys deletes the entities for the given keys.
func DSDeleteKeys(kind *types.Kind, keys []*types.Key) error {
	ctx := kind.Context
	dsKeys := toDSKeys(keys)

	ctx.Infof(LogDatastoreAction("deleting", "from", keys, kind.Name))

	return ndsDel(ctx, dsKeys)
}
