package hrd

import (
	"appengine/datastore"
	"fmt"
	"strings"
	"time"
)

// Query represents a datastore query.
type Query struct {
	err    *error
	coll   *Collection
	typeOf qryType
	logs   []string
	limit  int
	opts   *operationOpts
	qry    *datastore.Query
}

type qryType int

const (
	// normal query
	fullQry qryType = 1 + iota

	// only query projected fields
	projectQry

	// fetch keys first, then use batch get to only load uncached entities
	hybridQry
)

// newQuery creates a new Query for the passed collection.
// The collection's options are used as default options.
func newQuery(coll *Collection) (ret *Query) {
	return &Query{
		coll:   coll,
		limit:  -1,
		typeOf: hybridQry,
		opts:   defaultOperationOpts(),
		qry:    datastore.NewQuery(coll.name),
		logs:   []string{"KIND " + coll.name},
	}
}

func (qry *Query) clone() *Query {
	ret := *qry
	ret.opts = qry.opts.clone()
	if len(qry.logs) > 0 {
		ret.logs = make([]string, len(qry.logs))
		copy(ret.logs, qry.logs)
	}
	return &ret
}

// Opts returns a derivative Query with all passed-in options applied.
func (qry *Query) Opts(opts ...Opt) (ret *Query) {
	ret = qry.clone()
	ret.opts = ret.opts.Apply(opts...)
	return
}

// NoHybrid returns a derivative Query that will not run a hybrid query.
func (qry *Query) NoHybrid() *Query {
	return qry.Hybrid(false)
}

// Hybrid returns a derivative Query which will run as a hybrid or non-hybrid
// query depending on the passed-in argument.
func (qry *Query) Hybrid(enabled bool) (ret *Query) {
	ret = qry.clone()
	if enabled {
		if ret.typeOf == fullQry {
			ret.typeOf = hybridQry
		}
	} else {
		if ret.typeOf == hybridQry {
			ret.typeOf = fullQry
		}
	}
	return ret
}

// Limit returns a derivative Query that has a limit on the number
// of results returned. A negative value means unlimited.
func (qry *Query) Limit(limit int) (ret *Query) {
	ret = qry.clone()
	if limit > 0 {
		ret.log("LIMIT %v", limit)
	} else {
		limit = -1
		ret.log("NO LIMIT")
	}
	ret.qry = ret.qry.Limit(limit)
	ret.limit = limit
	return ret
}

// NoLimit returns a derivative Query that has no limit on the number
// of results returned.
func (qry *Query) NoLimit() (ret *Query) {
	return qry.Limit(-1)
}

// Ancestor returns a derivative Query with an ancestor filter.
// The ancestor should not be nil.
func (qry *Query) Ancestor(k *Key) (ret *Query) {
	ret = qry.clone()
	ret.log("ANCESTOR '%v'", k.IDString())
	ret.qry = ret.qry.Ancestor(k.Key)
	return ret
}

// Project returns a derivative Query that yields only the passed fields.
// It cannot be used in a keys-only query.
func (qry *Query) Project(s ...string) (ret *Query) {
	ret = qry.clone()
	ret.log("PROJECT '%v'", strings.Join(s, "', '"))
	ret.qry = ret.qry.Project(s...)
	ret.typeOf = projectQry
	return ret
}

// EventualConsistency returns a derivative query that returns eventually
// consistent results. It only has an effect on ancestor queries.
func (qry *Query) EventualConsistency() (ret *Query) {
	ret = qry.clone()
	ret.log("EVENTUAL CONSISTENCY")
	ret.qry = ret.qry.EventualConsistency()
	return ret
}

// End returns a derivative Query with the passed end point.
func (qry *Query) End(c string) (ret *Query) {
	ret = qry.clone()
	if c != "" {
		if cursor, err := datastore.DecodeCursor(c); err == nil {
			ret.log("END CURSOR")
			ret.qry = ret.qry.End(cursor)
		} else {
			err = fmt.Errorf("invalid end cursor (%v)", err)
			ret.err = &err
		}
	}
	return ret
}

// Start returns a derivative Query with the passed start point.
func (qry *Query) Start(c string) (ret *Query) {
	ret = qry.clone()
	if c != "" {
		if cursor, err := datastore.DecodeCursor(c); err == nil {
			ret.log("START CURSOR")
			ret.qry = ret.qry.Start(cursor)
		} else {
			err = fmt.Errorf("invalid start cursor (%v)", err)
			ret.err = &err
		}
	}
	return ret
}

// Offset returns a derivative Query that has an offset of how many keys
// to skip over before returning results. A negative value is invalid.
func (qry *Query) Offset(off int) (ret *Query) {
	ret = qry.clone()
	ret.log("OFFSET %v", off)
	ret.qry = ret.qry.Offset(off)
	return
}

// OrderAsc returns a derivative Query with a field-based sort order, ascending.
// Orders are applied in the order they are added.
func (qry *Query) OrderAsc(s string) (ret *Query) {
	ret = qry.clone()
	ret.log("ORDER ASC %v", s)
	ret.qry = ret.qry.Order(s)
	return ret
}

// OrderDesc returns a derivative Query with a field-based sort order, descending.
// Orders are applied in the order they are added.
func (qry *Query) OrderDesc(s string) (ret *Query) {
	ret = qry.clone()
	ret.log("ORDER DESC %v", s)
	ret.qry = ret.qry.Order("-" + s)
	return
}

// Filter returns a derivative Query with a field-based filter.
// The filterStr argument must be a field name followed by optional space,
// followed by an operator, one of ">", "<", ">=", "<=", or "=".
// Fields are compared against the provided value using the operator.
// Multiple filters are AND'ed together.
func (qry *Query) Filter(q string, val interface{}) (ret *Query) {
	ret = qry.clone()
	ret.log("FILTER '%v %v'", q, val)
	ret.qry = ret.qry.Filter(q, val)
	return
}

// ==== CACHE

// NoCache prevents reading/writing entities from/to
// the in-memory cache or memcache in this load operation.
func (qry *Query) NoCache() (ret *Query) {
	return qry.NoLocalCache().NoGlobalCache()
}

// NoLocalCache prevents reading/writing entities from/to
// the in-memory cache in this load operation.
func (qry *Query) NoLocalCache() (ret *Query) {
	return qry.NoLocalCacheWrite().NoLocalCacheRead()
}

// NoGlobalCache prevents reading/writing entities from/to
// memcache in this load operation.
func (qry *Query) NoGlobalCache() (ret *Query) {
	return qry.NoGlobalCacheWrite().NoGlobalCacheRead()
}

// CacheExpire sets the expiration time in memcache for entities
// that are cached after loading them to the datastore.
func (qry *Query) CacheExpire(exp time.Duration) (ret *Query) {
	q := qry.clone()
	q.opts = q.opts.CacheExpire(exp)
	return q
}

// NoCacheRead prevents reading entities from
// the in-memory cache or memcache in this load operation.
func (qry *Query) NoCacheRead() (ret *Query) {
	return qry.NoGlobalCacheRead().NoLocalCacheRead()
}

// NoLocalCacheRead prevents reading entities from
// the in-memory cache in this load operation.
func (qry *Query) NoLocalCacheRead() (ret *Query) {
	q := qry.clone()
	q.opts = q.opts.NoLocalCacheRead()
	return q
}

// NoGlobalCacheRead prevents reading entities from
// memcache in this load operation.
func (qry *Query) NoGlobalCacheRead() (ret *Query) {
	q := qry.clone()
	q.opts = q.opts.NoGlobalCacheRead()
	return q
}

// NoCacheWrite prevents writing entities to
// the in-memory cache or memcache in this load operation.
func (qry *Query) NoCacheWrite() (ret *Query) {
	return qry.NoGlobalCacheWrite().NoLocalCacheWrite()
}

// NoLocalCacheWrite prevents writing entities to
// the in-memory cache in this load operation.
func (qry *Query) NoLocalCacheWrite() (ret *Query) {
	q := qry.clone()
	q.opts = q.opts.NoLocalCacheWrite()
	return q
}

// NoGlobalCacheWrite prevents writing entities to
// memcache in this load operation.
func (qry *Query) NoGlobalCacheWrite() (ret *Query) {
	q := qry.clone()
	q.opts = q.opts.NoGlobalCacheWrite()
	return q
}

// ==== EXECUTE

// GetCount returns the number of results for the query.
func (qry *Query) GetCount() (int, error) {
	qry.log("COUNT")
	qry.coll.store.ctx.Infof(qry.getLog())

	if qry.err != nil {
		return 0, *qry.err
	}
	return qry.qry.Count(qry.coll.store.ctx)
}

// GetKeys executes the query as keys-only: No entities are retrieved, just their keys.
func (qry *Query) GetKeys() ([]*Key, string, error) {
	q := qry.clone()
	q.qry = q.qry.KeysOnly()
	q.log("KEYS-ONLY")

	it := q.Run()
	keys, err := it.GetAll(nil)
	if err != nil {
		return nil, "", err
	}
	cursor, err := it.Cursor()
	return keys, cursor, err
}

// GetAll runs the query and writes the entities to the passed destination.
//
// Note that, if not manually disabled, queries for more than 1 item use
// a "hybrid query". This means that first a keys-only query is executed
// and then the keys are used to lookup the local and global cache as well
// as the datastore eventually. For a warm cache this usually is
// faster and cheaper than the regular query.
func (qry *Query) GetAll(dsts interface{}) ([]*Key, string, error) {
	if qry.err != nil {
		return nil, "", *qry.err
	}

	if qry.limit != 1 && qry.typeOf == hybridQry && qry.opts.readGlobalCache {
		keys, cursor, err := qry.GetKeys()
		if err == nil && len(keys) > 0 {
			keys, err = newLoader(qry.coll).Keys(keys...).GetAll(dsts)
		}
		return keys, cursor, err
	}

	it := qry.Run()
	keys, err := it.GetAll(dsts)
	if err != nil {
		return nil, "", err
	}

	cursor, err := it.Cursor()
	return keys, cursor, err
}

// GetFirst executes the query and writes the result's first entity
// to the passed destination.
func (qry *Query) GetFirst(dst interface{}) (err error) {
	return qry.Run().GetOne(dst)
}

// Run executes the query and returns an Iterator.
func (qry *Query) Run() *Iterator {
	qry.coll.store.ctx.Infof(qry.getLog())
	return &Iterator{qry, qry.qry.Run(qry.coll.store.ctx)}
}

func (qry *Query) log(s string, values ...interface{}) {
	qry.logs = append(qry.logs, fmt.Sprintf(s, values...))
}

func (qry *Query) getLog() string {
	return fmt.Sprintf("running query \"%v\"", strings.Join(qry.logs, " | "))
}
