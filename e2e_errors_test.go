package hrd

import (
	. "github.com/101loops/bdd"
)

var _ = Describe("HRD Errors", func() {

	var (
		coll *Collection
	)

	BeforeEach(func() {
		clearCache()
		coll = store.Coll("A")
	})

	With("save", func() {

		It("does not save nil entity", func() {
			_, err := coll.Save().Entity(nil)

			Check(err, Contains, "must be non-nil")
		})

		It("does not save non-struct entity", func() {
			_, err := coll.Save().Entity(42)

			Check(err, Contains, "invalid value kind").And(Contains, "int")
		})

		It("does not save entity without ID() and SetID()", func() {
			invalidMdl := &InvalidModel{}
			_, err := coll.Save().Entity(invalidMdl)

			Check(err, Contains, "does not provide ID")
		})

		It("does not save complete entity without Id", func() {
			entity := &SimpleModel{}
			_, err := coll.Save().WithKey().Entity(entity)

			Check(err, Contains, "incomplete key")
		})
	})

	With("load", func() {

		It("does not load entity from struct", func() {
			var entity SimpleModel
			key, err := coll.Load().ID(42).GetOne(entity)

			Check(key, IsNil)
			Check(err, Contains, "invalid value kind").And(Contains, "struct")
		})

		It("does not load entity from non-reference struct", func() {
			var entity *SimpleModel
			key, err := coll.Load().ID(42).GetOne(entity)

			Check(key, IsNil)
			Check(err, Contains, "invalid value kind").And(Contains, "ptr")
		})

		It("does not load entities from map with invalid key", func() {
			var entities map[bool]*SimpleModel
			keys, err := coll.Load().IDs(1, 2, 3).GetAll(&entities)

			Check(keys, IsEmpty)
			Check(err, Contains, "invalid value key")
		})

		It("does not load non-existing entity", func() {
			var entity *SimpleModel
			key, err := coll.Load().ID(666).GetOne(&entity)

			Check(err, IsNil)
			Check(key, IsNil)
			Check(entity, IsNil)
		})
	})

	With("query", func() {

		It("does not run query with invalid cursor", func() {
			var entity *SimpleModel

			_, err := coll.Query().End("nonsense").GetCount()
			Check(err, Contains, "invalid end cursor")

			_, _, err = coll.Query().End("nonsense").GetKeys()
			Check(err, Contains, "invalid end cursor")

			err = coll.Query().End("nonsense").GetFirst(&entity)
			Check(err, Contains, "invalid end cursor")

			_, _, err = coll.Query().Start("nonsense").GetAll(&entity)
			Check(err, Contains, "invalid start cursor")

			_, _, err = coll.Query().Start("nonsense").NoHybrid().GetAll(&entity)
			Check(err, Contains, "invalid start cursor")
		})
	})
})
