package hrd

import (
	. "github.com/101loops/bdd"
	"github.com/101loops/hrd/internal"
	"github.com/101loops/hrd/internal/types"
)

var _ = Describe("Deleter", func() {

	BeforeEach(func() {
		dsDelete = func(_ *types.Kind, _ interface{}, _ bool) error {
			panic("unexpected call")
		}
		dsDeleteKeys = func(_ *types.Kind, _ ...*types.Key) error {
			panic("unexpected call")
		}
	})

	AfterEach(func() {
		dsDelete = internal.Delete
		dsDeleteKeys = internal.DeleteKeys
	})

	It("should delete an entity by key", func() {
		dsDeleteKeys = func(kind *types.Kind, keys ...*types.Key) error {
			Check(keys, Equals, toInternalKeys(myKind.NewNumKeys(42)))
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).Key(myKind.NewNumKey(42))
	})

	It("should delete multiple entities by key", func() {
		hrdKeys := []*Key{myKind.NewNumKey(1), myKind.NewNumKey(2)}

		dsDeleteKeys = func(kind *types.Kind, keys ...*types.Key) error {
			Check(keys, Equals, toInternalKeys(hrdKeys))
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).Keys(hrdKeys)
	})

	It("should delete an entity by numeric id", func() {
		dsDeleteKeys = func(kind *types.Kind, keys ...*types.Key) error {
			Check(keys, Equals, toInternalKeys(myKind.NewNumKeys(42)))
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).ID(42)
	})

	It("should delete multiple entities by numeric id", func() {
		dsDeleteKeys = func(kind *types.Kind, keys ...*types.Key) error {
			Check(keys, Equals, toInternalKeys(myKind.NewNumKeys(1, 2)))
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).IDs(1, 2)
	})

	It("should delete an entity by text id", func() {
		dsDeleteKeys = func(kind *types.Kind, keys ...*types.Key) error {
			Check(keys, Equals, toInternalKeys(myKind.NewTextKeys("a")))
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).TextID("a")
	})

	It("should delete multiple entities by text id", func() {
		dsDeleteKeys = func(kind *types.Kind, keys ...*types.Key) error {
			Check(keys, Equals, toInternalKeys(myKind.NewTextKeys("a", "z")))
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).TextIDs("a", "z")
	})

	It("should delete an entity", func() {
		entity := &MyModel{}

		dsDelete = func(kind *types.Kind, src interface{}, multi bool) error {
			Check(multi, IsFalse)
			Check(src, Equals, entity)
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).Entity(entity)
	})

	It("should delete multiple entities", func() {
		entities := []*MyModel{&MyModel{}, &MyModel{}}

		dsDelete = func(kind *types.Kind, srcs interface{}, multi bool) error {
			Check(multi, IsTrue)
			Check(srcs, Equals, entities)
			Check(kind.Name, Equals, "my-kind")
			return nil
		}

		myKind.Delete(ctx).Entities(entities)
	})
})
