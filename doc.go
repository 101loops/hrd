package hrd

import (
	"appengine/datastore"
	"fmt"
	"github.com/101loops/reflector"
	"reflect"
	"strings"
	"time"
)

type doc struct {
	src_val reflect.Value
	codec   *codec
	synced  bool
}

type property struct {
	name    string
	value   interface{}
	indexed bool
	multi   bool
}

func newDoc(src_val reflect.Value) (*doc, error) {
	src_type := src_val.Type()
	src_kind := src_val.Kind()
	switch src_kind {
	case reflect.Struct:
	case reflect.Ptr:
		src_type = src_val.Elem().Type()
		src_kind = src_val.Elem().Kind()
		if src_kind != reflect.Struct {
			return nil, fmt.Errorf("invalid value kind %q (wanted struct pointer)", src_kind)
		}
	default:
		return nil, fmt.Errorf("invalid value kind %q (wanted struct or struct pointer)", src_kind)
	}

	codec, err := getCodec(src_type)
	if err != nil {
		return nil, err
	}

	return &doc{src_val, codec, false}, nil
}

func newDocFromInst(src interface{}) (*doc, error) {
	return newDoc(reflect.ValueOf(src))
}

func newDocFromType(typ reflect.Type) (*doc, error) {
	return newDoc(reflect.New(typ.Elem()))
}

func (self *doc) nil() {
	dst := self.val()
	dst.Set(reflect.New(dst.Type()).Elem())
}

func (self *doc) get() interface{} {
	return self.src_val.Interface()
}

func (self *doc) set(src interface{}) {
	dst := self.val()
	v := reflect.ValueOf(src)
	if v.Kind() == reflect.Ptr && dst.Kind() != reflect.Ptr {
		v = v.Elem()
	}
	dst.Set(v)
}

func (self *doc) setKey(k *Key) {
	setKey(self.get(), k)
}

func (self *doc) val() reflect.Value {
	v := self.src_val
	if !v.CanSet() {
		v = self.src_val.Elem()
	}
	return v
}

func (self *doc) toProperties(prefix string, tags []string, multi bool) (res []*property, err error) {
	var props []*property

	src_val := self.val()
	for i, t := range self.codec.byIndex {
		v := src_val.Field(i)
		if !v.IsValid() || !v.CanSet() {
			continue
		}

		name := t.name
		if prefix != "" {
			name = prefix + "." + name
		}

		aggrTags := append(tags, t.tags...)

		// for slice fields (that aren't []byte), save each element
		if v.Kind() == reflect.Slice && v.Type() != typeOfByteSlice {
			for j := 0; j < v.Len(); j++ {
				props, err = itemToProperties(name, aggrTags, true, v.Index(j))
				if err != nil {
					return
				}
				res = append(res, props...)
			}
			continue
		}

		// otherwise, save the field itself
		props, err = itemToProperties(name, aggrTags, multi, v)
		if err != nil {
			return
		}
		res = append(res, props...)
	}

	return
}

// Note: Save should close the channel when done, even if an error occurred
func (self *doc) Save(c chan<- datastore.Property) error {
	defer close(c)

	src := self.get()

	// event: before save
	if hook, ok := src.(beforeSaver); ok {
		if err := hook.BeforeSave(); err != nil {
			return err
		}
	}

	// export properties
	props, err := self.toProperties("", []string{""}, false)
	if err != nil {
		return err
	}

	// fill channel
	for _, prop := range props {
		c <- datastore.Property{
			Name:     prop.name,
			Value:    prop.value,
			NoIndex:  !prop.indexed,
			Multiple: prop.multi,
		}
	}
	self.synced = true

	// event: after save
	if hook, ok := src.(afterSaver); ok {
		if err := hook.AfterSave(); err != nil {
			close(c)
			return err
		}
	}

	return nil
}

// Note: Load should drain the channel until closed, even if an error occurred
func (self *doc) Load(c <-chan datastore.Property) error {

	dst := self.get()

	// event: before load
	if hook, ok := dst.(beforeLoader); ok {
		if err := hook.BeforeLoad(); err != nil {
			return err
		}
	}

	if err := datastore.LoadStruct(dst, c); err != nil {
		return err
	}
	self.synced = true

	// event: after load
	if hook, ok := dst.(afterLoader); ok {
		if err := hook.AfterLoad(); err != nil {
			return err
		}
	}

	return nil
}

func itemToProperties(name string, tags []string, multi bool, v reflect.Value) (props []*property, err error) {

	// dereference pointers, ignore nil
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	// process tags
	indexed := false
	for _, tag := range tags {
		tag = strings.ToLower(tag)
		if tag == "omitempty" {
			if reflector.IsDefault(v.Interface()) {
				return // ignore complete field if empty
			}
		} else if strings.HasPrefix(tag, "index") {
			indexed = true
			if strings.HasSuffix(tag, ":omitempty") && reflector.IsDefault(v.Interface()) {
				indexed = false // ignore index if empty
			}
		} else if tag != "" {
			err = fmt.Errorf("unknown tag %q", tag)
			return
		}
	}

	p := &property{
		name:  name,
		multi: multi,
	}
	props = []*property{p}
	p.indexed = indexed

	// serialize
	switch x := v.Interface().(type) {
	//case *Key:
	//	p.value = x
	case time.Time:
		p.value = x
	case []byte:
		p.indexed = false
		p.value = x
	default:
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			p.value = v.Int()
		case reflect.Bool:
			p.value = v.Bool()
		case reflect.String:
			p.value = v.String()
		case reflect.Float32, reflect.Float64:
			p.value = v.Float()
		case reflect.Struct:
			if !v.CanAddr() {
				return nil, fmt.Errorf("unsupported property %q (unaddressable)", name)
			}
			sub, err := newDocFromInst(v.Addr().Interface())
			if err != nil {
				return nil, fmt.Errorf("unsupported property %q (%v)", name, err)
			}
			return sub.toProperties(name, tags, multi)
		}
	}

	if p.value == nil {
		err = fmt.Errorf("unsupported struct field type %q (unidentifiable)", v.Type())
	}

	return
}
