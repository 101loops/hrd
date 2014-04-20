package hrd

import (
	. "github.com/101loops/bdd"
)

var _ = Describe("HRD Read/Write", func() {

	With("default settings", func() {
		basicTests()
	})

	With("w/o cache", func() {
		basicTests(NoCache)
	})

	With("w/o local cache", func() {
		basicTests(NoLocalCache)
	})

	With("w/o global cache", func() {
		basicTests(NoGlobalCache)
	})
})

func basicTests(opts ...Opt) {

	var (
		genID  int64
		coll   *Collection
		loader *Loader
		entity *SimpleModel
	)

	BeforeEach(func() {
		if coll == nil {
			coll = randomColl()
		}

		entity = &SimpleModel{Text: "text", Data: []byte{1, 2, 3}}
		key, err := coll.Save().Entity(entity)

		Check(err, IsNil)
		Check(key, NotNil)
		genID = key.IntID()
		Check(genID, IsGreaterThan, 0)
		Check(entity.ID(), Equals, genID)
		Check(entity.lifecycle, Equals, "after-save")

		clearCache()
		loader = coll.Load().Opts(opts...)
	})

	It("saves an entity (with id)", func() {
		entity.SetID(42)
		key, err := coll.Save().ReqKey().Entity(entity)

		Check(err, IsNil)
		Check(key, NotNil)
		Check(key.IntID(), EqualsNum, 42)
		Check(entity.ID(), EqualsNum, 42)
	})

	It("loads an entity", func() {
		var entity *SimpleModel
		key, err := loader.ID(42).GetOne(&entity)

		Check(err, IsNil)
		Check(key, NotNil)
		Check(entity, NotNil)
		Check(entity.ID(), EqualsNum, 42)
		Check(entity.Text, Equals, "text")
		Check(entity.lifecycle, Equals, "after-load")
	})

	It("loads all entities into slice of struct pointers", func() {
		var entities []*SimpleModel
		keys, err := coll.Load().IDs(1, 42, genID).GetAll(&entities)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)

		Check(keys[0].IntID(), EqualsNum, 1)
		Check(keys[0].source, Equals, "")
		Check(keys[0].Exists(), IsFalse)

		Check(keys[1].IntID(), EqualsNum, 42)
		Check(keys[1].source, Equals, sourceDatastore)
		Check(entities[1].Text, Equals, "text")

		Check(keys[2].IntID(), EqualsNum, genID)
		Check(keys[2].source, Equals, sourceDatastore)
		Check(entities[2].Text, Equals, "text")
	})

	It("loads all entities into map of type Key -> struct pointers", func() {
		var entities map[*Key]*SimpleModel
		keys, err := coll.Load().IDs(1, 42, genID).GetAll(&entities)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)
	})

	It("loads all entities into map of type int64 -> struct pointers", func() {
		var entities map[int64]*SimpleModel
		keys, err := coll.Load().IDs(1, 42, genID).GetAll(&entities)

		Check(err, IsNil)
		Check(keys, HasLen, 3)
		Check(entities, HasLen, 3)
	})

	It("deletes an entity", func() {
		err := coll.Delete().ID(42)

		Check(err, IsNil)

		var entity *SimpleModel
		key, err := loader.ID(42).GetOne(&entity)

		Check(err, IsNil)
		Check(key, IsNil)
		Check(entity, IsNil)
	})
}
