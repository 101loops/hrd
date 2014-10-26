package hrd

import "appengine/datastore"

const deleteMultiLimit = 500

func (store *Store) deleteMulti(kind string, keys []*Key) (err error) {

	store.ctx.Infof(store.logAct("deleting", "from", keys, kind))

	// #1 delete from cache
	defer store.cache.delete(keys)

	// #2 delete from datastore
	for i := 0; i <= len(keys)/deleteMultiLimit; i++ {
		lo := i * deleteMultiLimit
		hi := (i + 1) * deleteMultiLimit
		if hi > len(keys) {
			hi = len(keys)
		}
		err = datastore.DeleteMulti(store.ctx, toDSKeys(keys[lo:hi]))
		// TODO: appengine.MultiError
		if err != nil {
			return err
		}
	}

	return
}
