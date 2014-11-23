package internal

import (
	. "github.com/101loops/bdd"
	"github.com/101loops/hrd/internal/types"

	ds "appengine/datastore"
)

var _ = Describe("Get", func() {

	With("w/ global cache", func() {
		dsLoadTests(true)
	})

	With("w/o global cache", func() {
		dsLoadTests(false)
	})
})

func dsLoadTests(useGlobalCache bool) {

	var (
		kind *types.Kind
	)

	BeforeEach(func() {
		kind = randomKind()

		entities := make([]interface{}, 4)
		for i := int64(1); i < 5; i++ {
			entity := &MyModel{Num: i}
			entity.SetID(i)
			entities[i-1] = entity
		}
		keys, err := Put(kind, entities, true)
		Check(err, IsNil)
		Check(keys, HasLen, 4)

		clearCache()
	})

	It("loads an entity", func() {
		var entity *MyModel
		dsKey := ds.NewKey(ctx, kind.Name, "", 1, nil)
		keys, err := Get(kind, types.NewKeys(dsKey), &entity, useGlobalCache, false)

		Check(err, IsNil)
		Check(keys, HasLen, 1)
		Check(keys[0].Synced, NotNil)
		Check(entity.ID(), EqualsNum, 1)
		Check(entity.Num, EqualsNum, 1)
	})

	It("loads multiple entities into slice of struct pointers", func() {
		var entities []*MyModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name, "", 1, nil),
			ds.NewKey(ctx, kind.Name, "", 2, nil),
			ds.NewKey(ctx, kind.Name, "", 666, nil),
		}
		keys, err := Get(kind, types.NewKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)

		Check(keys[0].IntID(), EqualsNum, 1)
		Check(keys[0].Synced, NotNil)
		Check(entities[0], NotNil)

		Check(keys[1].IntID(), EqualsNum, 2)
		Check(keys[1].Synced, NotNil)
		Check(entities[1], NotNil)

		Check(keys[2].IntID(), EqualsNum, 666)
		Check(keys[2].Synced, IsNil)
		// Check(entities[2], IsNil) TODO
	})

	It("loads multiple entities into map of struct pointers by Key", func() {
		var entities map[*types.Key]*MyModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name, "", 1, nil),
			ds.NewKey(ctx, kind.Name, "", 2, nil),
			ds.NewKey(ctx, kind.Name, "", 666, nil),
		}
		keys, err := Get(kind, types.NewKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)
	})

	It("loads multiple entities into map of struct pointers by int64", func() {
		var entities map[int64]*MyModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name, "", 1, nil),
			ds.NewKey(ctx, kind.Name, "", 2, nil),
			ds.NewKey(ctx, kind.Name, "", 666, nil),
		}
		keys, err := Get(kind, types.NewKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)
	})

	// ==== ERRORS

	It("does not load entity into invalid type", func() {
		var entity string
		dsKey := ds.NewKey(ctx, kind.Name, "", 1, nil)
		keys, err := Get(kind, types.NewKeys(dsKey), entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And("invalid value kind").And(Contains, "string")
	})

	It("does not load entity into non-pointer struct", func() {
		var entity MyModel
		dsKey := ds.NewKey(ctx, kind.Name, "", 1, nil)
		keys, err := Get(kind, types.NewKeys(dsKey), entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "invalid value kind").And(Contains, "struct")
	})

	It("does not load entity into non-reference struct", func() {
		var entity *MyModel
		dsKey := ds.NewKey(ctx, kind.Name, "", 1, nil)
		keys, err := Get(kind, types.NewKeys(dsKey), entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "invalid value kind").And(Contains, "ptr")
	})

	It("does not load entities into map with invalid key", func() {
		var entities map[bool]*MyModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name, "", 1, nil),
			ds.NewKey(ctx, kind.Name, "", 2, nil),
		}
		keys, err := Get(kind, types.NewKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "invalid value key")
	})

	It("does not accept key for different Kind", func() {
		var entity *MyModel
		invalidKey := ds.NewKey(ctx, "wrong-kind", "", 1, nil)
		keys, err := Get(kind, types.NewKeys(invalidKey), &entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(entity, IsNil)
		Check(err, NotNil).And(Contains, "invalid key kind 'wrong-kind'")
	})

	It("does not load empty keys", func() {
		var entities []*MyModel
		keys, err := Get(kind, nil, &entities, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "no keys provided")
	})

	It("does not load incomplete key", func() {
		var entity *MyModel
		incompleteKey := ds.NewKey(ctx, kind.Name, "", 0, nil)
		keys, err := Get(kind, types.NewKeys(incompleteKey), &entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "is incomplete")
	})
}
