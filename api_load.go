package hrd

import (
	"fmt"
)

// Loader can load entities from a Collection.
type Loader struct {
	coll *Collection
	opts *operationOpts
	keys []*Key
}

// SingleLoader is a special Loader that allows to fetch exactly one entity
// from the datastore.
type SingleLoader struct {
	loader *Loader
}

// MultiLoader is a special Loader that allows to fetch multiple entities
// from the datastore.
type MultiLoader struct {
	loader *Loader
}

// newLoader creates a new Loader for the passed collection.
// The collection's options are used as default options.
func newLoader(coll *Collection) *Loader {
	return &Loader{coll: coll, opts: coll.opts.clone()}
}

// Key loads a single entity by key from the datastore.
func (l *Loader) Key(key *Key) *SingleLoader {
	l.keys = []*Key{key}
	return &SingleLoader{l}
}

// Keys load multiple entities by key from the datastore.
func (l *Loader) Keys(keys ...*Key) *MultiLoader {
	l.keys = keys
	return &MultiLoader{l}
}

// ID loads a single entity by id from the datastore.
func (l *Loader) ID(id int64, parent ...*Key) *SingleLoader {
	l.keys = []*Key{l.coll.NewNumKey(id, parent...)}
	return &SingleLoader{l}
}

// TextID loads a single key by text id from the datastore.
func (l *Loader) TextID(id string, parent ...*Key) *SingleLoader {
	l.keys = []*Key{l.coll.NewTextKey(id, parent...)}
	return &SingleLoader{l}
}

// IDs load multiple keys by id from the datastore.
func (l *Loader) IDs(ids ...int64) *MultiLoader {
	l.keys = l.coll.NewNumKeys(ids...)
	return &MultiLoader{l}
}

// TextIDs load multiple keys by text id from the datastore.
func (l *Loader) TextIDs(ids ...string) *MultiLoader {
	l.keys = l.coll.NewTextKeys(ids...)
	return &MultiLoader{l}
}

// ==== CONFIG

// Opts applies a sequence of Opt the Loader's options.
func (l *Loader) Opts(opts ...Opt) *Loader {
	l.opts = l.opts.Apply(opts...)
	return l
}

// GlobalCache defines whether entities are read from memcache.
// If no parameter is passed, true is assumed.
func (l *Loader) GlobalCache(enable ...bool) *Loader {
	l.opts = l.opts.GlobalCache(enable...)
	return l
}

// ==== EXECUTE

// TODO: func (l *Loader) GetEntity(dst interface{}) ([]*Key, error)
// TODO: func (l *Loader) GetEntities(dsts interface{}) ([]*Key, error)

// GetAll loads entities from the datastore into the passed destination.
func (l *MultiLoader) GetAll(dsts interface{}) ([]*Key, error) {
	return l.loader.get(dsts, true)
}

// GetOne loads an entity from the datastore into the passed destination.
func (l *SingleLoader) GetOne(dst interface{}) (*Key, error) {
	keys, err := l.loader.get(dst, false)

	var key *Key
	if len(keys) == 1 {
		if keys[0].Exists() {
			key = keys[0]
		}
	}

	return key, err
}

func (l *Loader) get(dst interface{}, multi bool) ([]*Key, error) {
	for _, k := range l.keys {
		keyKind := k.Kind()
		collKind := l.coll.name
		if keyKind != collKind {
			err := fmt.Errorf("invalid key kind '%v' for collection '%v'", keyKind, collKind)
			return nil, l.coll.store.logErr(err)
		}
	}

	docs, err := newWriteableDocs(dst, l.keys, multi)
	if err != nil {
		return nil, err
	}
	return l.coll.store.getMulti(l.coll.name, docs, l.opts)
}
