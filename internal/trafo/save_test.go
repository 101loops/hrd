package trafo

import (
	"fmt"
	"time"
	. "github.com/101loops/bdd"
	"github.com/101loops/hrd/entity"
	"github.com/101loops/hrd/internal/types"

	ae "appengine"
	ds "appengine/datastore"
)

type saveEntity struct {
	A string
	B int

	beforeFunc func() error
	afterFunc  func() error
}

var _ = Describe("Doc Save", func() {

	Context("fields", func() {

		It("should serialize primitives", func() {
			type MyModel struct {
				I   int
				I8  int8
				I16 int16
				I32 int32
				I64 int64
				B   bool
				S   string
				F32 float32
				F64 float64
			}

			props, err := save(&MyModel{
				int(1), int8(2), int16(3), int32(4), int64(5), true, "test", float32(1.0), float64(2.0),
			})
			Check(err, IsNil)
			Check(props, NotNil).And(HasLen, 9)

			Check(*props[0], Equals, ds.Property{"I", int64(1), true, false})
			Check(*props[1], Equals, ds.Property{"I8", int64(2), true, false})
			Check(*props[2], Equals, ds.Property{"I16", int64(3), true, false})
			Check(*props[3], Equals, ds.Property{"I32", int64(4), true, false})
			Check(*props[4], Equals, ds.Property{"I64", int64(5), true, false})
			Check(*props[5], Equals, ds.Property{"B", true, true, false})
			Check(*props[6], Equals, ds.Property{"S", "test", true, false})
			Check(*props[7], Equals, ds.Property{"F32", float64(1.0), true, false})
			Check(*props[8], Equals, ds.Property{"F64", float64(2.0), true, false})
		})

		It("should serialize known complex types", func() {
			type MyModel struct {
				B  []byte
				T  time.Time
				K  *ds.Key
				KC types.DSKeyConverter
				BK ae.BlobKey
				GP ae.GeoPoint
			}

			dsKey := ds.NewKey(ctx, "kind", "", 42, nil)
			entity := &MyModel{
				[]byte("test"), time.Now(),
				dsKey, types.ImportKey(dsKey),
				ae.BlobKey("bkey"), ae.GeoPoint{1, 2},
			}
			props, err := save(entity)
			Check(err, IsNil)
			Check(props, NotNil).And(HasLen, 6)

			Check(*props[0], Equals, ds.Property{"B", entity.B, true, false})
			Check(*props[1], Equals, ds.Property{"T", entity.T, true, false})
			Check(*props[2], Equals, ds.Property{"K", entity.K, true, false})
			Check(*props[3], Equals, ds.Property{"KC", entity.K, true, false})
			Check(*props[4], Equals, ds.Property{"BK", entity.BK, true, false})
			Check(*props[5], Equals, ds.Property{"GP", entity.GP, true, false})
		})

		It("should serialize arbitrary complex fields", func() {
			type Pair struct {
				Key string `datastore:"key"`
				Val string
			}

			type MyModel struct {
				Struct Pair   `datastore:"tag"`
				Slice  []Pair `datastore:"tags"`
			}

			props, err := save(&MyModel{
				Struct: Pair{"life", "42"},
				Slice:  []Pair{Pair{"Bill", "Bob"}, Pair{"Barb", "Betty"}},
			})
			Check(err, IsNil)
			Check(props, NotNil)
			Check(props, HasLen, 6)

			Check(*props[0], Equals, ds.Property{"tag.key", "life", true, false})
			Check(*props[1], Equals, ds.Property{"tag.Val", "42", true, false})
			Check(*props[2], Equals, ds.Property{"tags.key", "Bill", true, true})
			Check(*props[3], Equals, ds.Property{"tags.Val", "Bob", true, true})
			Check(*props[4], Equals, ds.Property{"tags.key", "Barb", true, true})
			Check(*props[5], Equals, ds.Property{"tags.Val", "Betty", true, true})
		})

		It("should serialize embedded fields", func() {
			type Embedded1 struct {
				Data string
			}
			type Embedded2 struct {
				Data string
			}
			type MyModel struct {
				Embedded1
				Embedded2 `datastore:"embedded"`
			}

			props, err := save(&MyModel{})
			Check(err, IsNil)
			Check(props, NotNil).And(HasLen, 2)
			Check(*props[0], Equals, ds.Property{"Data", "", true, false})
			Check(*props[1], Equals, ds.Property{"embedded.Data", "", true, false})
		})

		It("should skip ignored fields", func() {
			type MyModel struct {
				Ignored string `datastore:"-"`
			}

			props, err := save(&MyModel{})
			Check(err, IsNil)
			Check(props, IsEmpty)
		})

		// ==== ERRORS

		It("should report unsupported field type", func() {
			type InvalidModel struct {
				Complex128 complex128
			}
			_, err := save(&InvalidModel{})
			Check(err, Contains, `unsupported struct field type "complex128"`)
		})

		It("should report unsupported field type in nested struct", func() {
			type InvalidModel struct {
				Complex128 complex128
			}
			type MyModel struct {
				InvalidModel
			}

			_, err := save(&MyModel{InvalidModel{}})
			Check(err, Contains, `unsupported struct field type "complex128"`)
		})
	})

	Context("tags", func() {

		It("should omit empty fields", func() {
			type MyModel struct {
				Bool    bool      `datastore:",omitempty"`
				Integer int64     `datastore:",omitempty"`
				String  string    `datastore:",omitempty"`
				Time    time.Time `datastore:",omitempty"`
				Bytes   []byte    `datastore:",omitempty"`
			}
			props, err := save(&MyModel{})

			Check(err, IsNil)
			Check(props, NotNil).And(HasLen, 0)
		})

		It("should index fields", func() {
			type MyModel struct {
				Field string `datastore:",index"`
				Empty string `datastore:",index:omitempty"`
			}
			props, err := save(&MyModel{"something", ""})

			Check(err, IsNil)
			Check(props, NotNil).And(HasLen, 2)
			Check(*props[0], Equals, ds.Property{"Field", "something", false, false})
			Check(*props[1], Equals, ds.Property{"Empty", "", true, false})
		})

		// ==== ERRORS

		It("should report invalid tag", func() {
			type MyModel struct {
				Field string `datastore:",invalid-tag"`
			}
			props, err := save(&MyModel{})

			Check(props, IsEmpty)
			Check(err, Contains, `unknown tag "invalid-tag"`)
		})
	})

	Context("timestamp", func() {

		var now time.Time

		BeforeEach(func() {
			now = time.Now()
			nowFunc = func() time.Time {
				return now
			}
		})

		It("should set created at", func() {
			type MyModel struct {
				entity.CreatedTime
			}
			entity := &MyModel{}
			props, err := save(entity)

			Check(err, IsNil)
			Check(props, HasLen, 1)
			Check(*props[0], Equals, ds.Property{"created_at", now, false, false})
		})

		It("should set updated at", func() {
			type MyModel struct {
				entity.UpdatedTime
			}
			entity := &MyModel{}
			props, err := save(entity)

			Check(err, IsNil)
			Check(props, HasLen, 1)
			Check(*props[0], Equals, ds.Property{"updated_at", now, false, false})
		})
	})

	Context("lifecycle hooks", func() {

		It("should run lifecycle hooks", func() {
			var hooks []string
			entity := &HookEntity{}
			entity.beforeSave = func() error {
				hooks = append(hooks, "before")
				return nil
			}
			entity.afterSave = func() error {
				hooks = append(hooks, "after")
				return nil
			}

			_, err := save(entity)
			Check(err, IsNil)
			Check(hooks, Equals, []string{"before", "after"})
		})

		// ==== ERRORS

		It("should return an error when BeforeSave fails", func() {
			entity := &HookEntity{}
			entity.beforeSave = func() error {
				return fmt.Errorf("an error")
			}

			_, err := save(entity)
			Check(err, HasOccurred)
		})

		It("should return an error when AfterSave fails", func() {
			entity := &HookEntity{}
			entity.afterSave = func() error {
				return fmt.Errorf("an error")
			}

			_, err := save(entity)
			Check(err, HasOccurred)
		})
	})
})

func save(src interface{}) ([]*ds.Property, error) {
	CodecSet.AddMust(src)

	doc, err := newDocFromInst(src)
	if err != nil {
		panic(err)
	}

	return doc.Save(ctx)
}
