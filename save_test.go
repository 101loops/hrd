package hrd

import (
	. "github.com/101loops/bdd"
	"github.com/101loops/hrd/internal"
	"github.com/101loops/hrd/internal/types"
)

var _ = Describe("Saver", func() {

	BeforeEach(func() {
		dsPut = func(_ *types.Kind, _ interface{}, _ bool) ([]*types.Key, error) {
			panic("unexpected call")
		}
	})

	AfterEach(func() {
		dsPut = internal.Put
	})

	It("saves an entity", func() {
		entity := &MyModel{}

		dsPut = func(kind *types.Kind, src interface{}, completeKeys bool) ([]*types.Key, error) {
			// TODO
			Check(completeKeys, IsFalse)
			Check(kind.Name, Equals, "my-kind")
			return newNumKeys(42), nil
		}

		key, err := myKind.Save(ctx).Entity(entity)
		Check(err, IsNil)
		Check(key, Equals, myKind.NewNumKey(42))
	})

	It("saves multiple entities", func() {
		entities := []*MyModel{&MyModel{}, &MyModel{}}

		dsPut = func(kind *types.Kind, src interface{}, completeKeys bool) ([]*types.Key, error) {
			// TODO
			Check(completeKeys, IsFalse)
			Check(kind.Name, Equals, "my-kind")
			return newNumKeys(1, 2), nil
		}

		keys, err := myKind.Save(ctx).Entities(entities)
		Check(err, IsNil)
		Check(keys, Equals, myKind.NewNumKeys(1, 2))
	})

	It("can require complete keys", func() {
		dsPut = func(_ *types.Kind, _ interface{}, completeKeys bool) ([]*types.Key, error) {
			Check(completeKeys, IsTrue)
			return nil, nil
		}

		myKind.Save(ctx).Opts(CompleteKeys).Entity(&MyModel{})
	})
})