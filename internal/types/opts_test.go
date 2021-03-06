package types

import (
	. "github.com/101loops/bdd"
)

var _ = Describe("Options", func() {

	var opts *Opts

	BeforeEach(func() {
		opts = DefaultOpts()
	})

	It("should have default options", func() {
		Check(opts.CompleteKeys, IsFalse)
		Check(opts.NoGlobalCache, IsFalse)
	})

	It("should return clone", func() {
		clone := opts.Clone()
		Check(opts, Equals, clone)

		opts.CompleteKeys = true
		Check(opts, Not(Equals), clone)
	})
})
