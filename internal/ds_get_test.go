package internal

import (
	. "github.com/101loops/bdd"

	ds "appengine/datastore"
)

var _ = Describe("DSGet", func() {

	With("w/ global cache", func() {
		dsLoadTests(true)
	})

	With("w/o global cache", func() {
		dsLoadTests(false)
	})
})

func dsLoadTests(useGlobalCache bool) {

	var (
		kind Kind
	)

	BeforeEach(func() {
		kind = randomKind()

		entities := make([]interface{}, 4)
		for i := int64(0); i < 4; i++ {
			entity := &SimpleModel{}
			entity.SetID(i + 1)
			entities[i] = entity
		}
		keys, err := DSPut(kind, entities, true)
		Check(err, IsNil)
		Check(keys, HasLen, 4)

		clearCache()
	})

	It("loads an entity", func() {
		var entity *SimpleModel
		dsKey := ds.NewKey(ctx, kind.Name(), "", 1, nil)
		keys, err := DSGet(kind, newKeys(dsKey), &entity, useGlobalCache, false)

		Check(err, IsNil)
		Check(keys, HasLen, 1)
		Check(entity.ID(), EqualsNum, 1)
	})

	It("loads multiple entities into slice of struct pointers", func() {
		var entities []*SimpleModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name(), "", 1, nil),
			ds.NewKey(ctx, kind.Name(), "", 2, nil),
			ds.NewKey(ctx, kind.Name(), "", 666, nil),
		}
		keys, err := DSGet(kind, newKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)

		Check(keys[0].IntID(), EqualsNum, 1)
		Check(keys[0].Exists(), IsTrue)
		Check(keys[1].IntID(), EqualsNum, 2)
		Check(keys[1].Exists(), IsTrue)
		Check(keys[2].IntID(), EqualsNum, 666)
		Check(keys[2].Exists(), IsFalse)
	})

	It("loads multiple entities into map of struct pointers by Key", func() {
		var entities map[*Key]*SimpleModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name(), "", 1, nil),
			ds.NewKey(ctx, kind.Name(), "", 2, nil),
			ds.NewKey(ctx, kind.Name(), "", 666, nil),
		}
		keys, err := DSGet(kind, newKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)
	})

	It("loads multiple entities into map of struct pointers by int64", func() {
		var entities map[int64]*SimpleModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name(), "", 1, nil),
			ds.NewKey(ctx, kind.Name(), "", 2, nil),
			ds.NewKey(ctx, kind.Name(), "", 666, nil),
		}
		keys, err := DSGet(kind, newKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)
	})

	// ==== ERRORS

	It("does not load entity into invalid type", func() {
		var entity string
		dsKey := ds.NewKey(ctx, kind.Name(), "", 1, nil)
		keys, err := DSGet(kind, newKeys(dsKey), entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And("invalid value kind").And(Contains, "string")
	})

	It("does not load entity into non-pointer struct", func() {
		var entity SimpleModel
		dsKey := ds.NewKey(ctx, kind.Name(), "", 1, nil)
		keys, err := DSGet(kind, newKeys(dsKey), entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "invalid value kind").And(Contains, "struct")
	})

	It("does not load entity into non-reference struct", func() {
		var entity *SimpleModel
		dsKey := ds.NewKey(ctx, kind.Name(), "", 1, nil)
		keys, err := DSGet(kind, newKeys(dsKey), entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "invalid value kind").And(Contains, "ptr")
	})

	It("does not load entities into map with invalid key", func() {
		var entities map[bool]*SimpleModel
		dsKeys := []*ds.Key{
			ds.NewKey(ctx, kind.Name(), "", 1, nil),
			ds.NewKey(ctx, kind.Name(), "", 2, nil),
		}
		keys, err := DSGet(kind, newKeys(dsKeys...), &entities, useGlobalCache, true)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "invalid value key")
	})

	It("does not accept key for different Kind", func() {
		var entity *SimpleModel
		invalidKey := ds.NewKey(ctx, "wrong-kind", "", 1, nil)
		keys, err := DSGet(kind, newKeys(invalidKey), &entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(entity, IsNil)
		Check(err, NotNil).And(Contains, "invalid key kind 'wrong-kind'")
	})

	It("does not load empty keys", func() {
		var entities []*SimpleModel
		keys, err := DSGet(kind, nil, &entities, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "no keys provided")
	})

	It("does not load incomplete key", func() {
		var entity *SimpleModel
		incompleteKey := ds.NewKey(ctx, kind.Name(), "", 0, nil)
		keys, err := DSGet(kind, newKeys(incompleteKey), &entity, useGlobalCache, false)

		Check(keys, IsNil)
		Check(err, NotNil).And(Contains, "is incomplete")
	})
}
