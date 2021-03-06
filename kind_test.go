package hrd

import . "github.com/101loops/bdd"

var _ = Describe("Kind", func() {

	It("should create numeric key", func() {
		key := myKind.NewNumKey(42)

		Check(key.IntID(), EqualsNum, 42)
		Check(key.Parent(), IsNil)
	})

	It("should create numeric keys", func() {
		keys := myKind.NewNumKeys(1, 2)

		Check(keys, HasLen, 2)
		Check(keys[0].IntID(), EqualsNum, 1)
		Check(keys[1].IntID(), EqualsNum, 2)
	})

	It("should create numeric key with parent", func() {
		key := myKind.NewNumKey(42, myKind.NewNumKey(66))

		Check(key.IntID(), EqualsNum, 42)
		Check(key.Parent(), NotNil)
		Check(key.Parent().IntID(), EqualsNum, 66)
	})

	It("should create text key", func() {
		key := myKind.NewTextKey("abc")

		Check(key.StringID(), Equals, "abc")
		Check(key.Parent(), IsNil)
	})

	It("should create text keys", func() {
		keys := myKind.NewTextKeys("abc", "xyz")

		Check(keys, HasLen, 2)
		Check(keys[0].StringID(), Equals, "abc")
		Check(keys[1].StringID(), Equals, "xyz")
	})

	It("should create text key with parent", func() {
		key := myKind.NewTextKey("abc", myKind.NewTextKey("xyz"))

		Check(key.StringID(), Equals, "abc")
		Check(key.Parent(), NotNil)
		Check(key.Parent().StringID(), Equals, "xyz")
	})
})
