package internal

import (
	"fmt"

	"github.com/qedus/nds"

	ae "appengine"
	ds "appengine/datastore"
)

var (
	dsGet = func(ctx ae.Context, keys []*ds.Key, dst interface{}) error {
		return ds.GetMulti(ctx, keys, dst)
	}

	ndsGet = func(ctx ae.Context, keys []*ds.Key, dst interface{}) error {
		return nds.GetMulti(ctx, keys, dst)
	}
)

// DSGet loads entities for the given keys.
func DSGet(kind Kind, keys []*Key, dst interface{}, useGlobalCache bool, multi bool) ([]*Key, error) {
	if err := validateDSGetKeys(kind, keys); err != nil {
		return nil, err
	}

	docs, err := newWriteableDocs(dst, keys, multi)
	if err != nil {
		return nil, err
	}
	dsDocs := docs.list

	ctx := kind.Context()
	ctx.Infof(LogDatastoreAction("getting", "from", keys, kind.Name()))

	var dsErr error
	dsKeys := toDSKeys(keys)
	if useGlobalCache {
		dsErr = ndsGet(ctx, dsKeys, dsDocs)
	}
	dsErr = dsGet(ctx, dsKeys, dsDocs)

	return applyResult(dsDocs, dsKeys, dsErr)
}

func validateDSGetKeys(kind Kind, keys []*Key) error {
	if keys == nil || len(keys) == 0 {
		return fmt.Errorf("no keys provided")
	}

	for i, key := range keys {
		if key.Incomplete() {
			return fmt.Errorf("'%v' is incomplete (%dth index)", key, i)
		}
	}

	for _, k := range keys {
		keyKind := k.Kind()
		if keyKind != kind.Name() {
			err := fmt.Errorf("invalid key kind '%v' for Kind '%v'", keyKind, kind.Name())
			return logErr(kind.Context(), err)
		}
	}

	return nil
}
