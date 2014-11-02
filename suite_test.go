package hrd

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"testing"
	"time"
	. "github.com/101loops/bdd"
	"appengine/aetest"
	"appengine/memcache"
)

var (
	ctx   aetest.Context
	store *Store
)

func TestSuite(t *testing.T) {
	var err error
	ctx, err = aetest.NewContext(nil)
	if err != nil {
		panic(err)
	}
	defer ctx.Close()

	store = NewStore(ctx)

	RegisterEntityMust(SimpleModel{})
	RegisterEntityMust(InvalidModel{})
	RegisterEntityMust(ComplexModel{})

	RunSpecs(t, "HRD Suite")
}

func randomColl() *Collection {
	var n int32
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return store.Coll(fmt.Sprintf("coll_%v", n))
}

func clearCache() {
	memcache.Flush(ctx)
}

type InvalidModel struct{}

type SimpleModel struct {
	id        int64
	Ignore    string    `datastore:"-"`
	Num       int64     `datastore:"num"`
	Data      []byte    `datastore:",index"`
	Text      string    `datastore:"html,index"`
	Time      time.Time `datastore:"timing,index,omitempty"`
	lifecycle []string
	updatedAt time.Time
	createdAt time.Time
}

func (mdl *SimpleModel) ID() int64 {
	return mdl.id
}

func (mdl *SimpleModel) SetID(id int64) {
	mdl.id = id
}

func (mdl *SimpleModel) BeforeLoad() error {
	mdl.lifecycle = append(mdl.lifecycle, "before-load")
	return nil
}

func (mdl *SimpleModel) AfterLoad() error {
	mdl.lifecycle = append(mdl.lifecycle, "after-load")
	return nil
}

func (mdl *SimpleModel) BeforeSave() error {
	mdl.lifecycle = append(mdl.lifecycle, "before-save")
	return nil
}

func (mdl *SimpleModel) AfterSave() error {
	mdl.lifecycle = append(mdl.lifecycle, "after-save")
	return nil
}

func (mdl *SimpleModel) SetCreatedAt(t time.Time) {
	mdl.createdAt = t
}

func (mdl *SimpleModel) SetUpdatedAt(t time.Time) {
	mdl.updatedAt = t
}

type ComplexModel struct {
	Pair Pair `datastore:"tag"`
	//PairPtr  *Pair   `datastore:"pair"`
	Pairs []Pair `datastore:"tags"`
	//PairPtrs []*Pair `datastore:"pairs"`
	lifecycle string `datastore:"-"`
}

func (mdl *ComplexModel) BeforeLoad() error {
	mdl.lifecycle = "before-load"
	return nil
}

func (mdl *ComplexModel) AfterLoad() error {
	mdl.lifecycle = "after-load"
	return nil
}

func (mdl *ComplexModel) BeforeSave() error {
	mdl.lifecycle = "before-save"
	return nil
}

func (mdl *ComplexModel) AfterSave() error {
	mdl.lifecycle = "after-save"
	return nil
}

type Pair struct {
	Key string `datastore:"key,index,omitempty"`
	Val string
}
