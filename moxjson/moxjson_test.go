package moxjson_test

import (
	"encoding/json"
	"github.com/lithictech/moxpopuli/moxjson"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestMoxjson(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "moxjson Suite")
}

var _ = Describe("moxjson", func() {
	It("parses paths", func() {
		Expect(moxjson.ParsePath("").String()).To(Equal(""))
		Expect(moxjson.ParsePath("1").String()).To(Equal("1"))
		Expect(moxjson.ParsePath("x.1").String()).To(Equal("x.1"))
		Expect(moxjson.ParsePath("x.[1]").String()).To(Equal("x.1"))
	})
	It("can get and set paths in JSON", func() {
		js := `{"x": 1, "arr": [1, [9, 10], [{"a": 5}]]}`
		var j interface{}
		Expect(json.Unmarshal([]byte(js), &j)).To(Succeed())
		Expect(moxjson.Get(j, moxjson.ParsePath("arr.0"))).To(BeEquivalentTo(1))
		Expect(moxjson.Get(j, moxjson.ParsePath("arr.[0]"))).To(BeEquivalentTo(1))
		Expect(moxjson.Get(j, moxjson.ParsePath("x"))).To(BeEquivalentTo(1))
		Expect(moxjson.Get(j, moxjson.ParsePath("arr.1.1"))).To(BeEquivalentTo(10))
		Expect(moxjson.Get(j, moxjson.ParsePath("arr.[1].1"))).To(BeEquivalentTo(10))
		Expect(moxjson.Get(j, moxjson.ParsePath("arr.2.0.a"))).To(BeEquivalentTo(5))

		Expect(moxjson.Set(j, 5, moxjson.ParsePath("arr.0"))).To(Succeed())
		Expect(moxjson.Get(j, moxjson.ParsePath("arr.0"))).To(BeEquivalentTo(5))
		Expect(moxjson.Set(j, 6, moxjson.ParsePath("y"))).To(Succeed())
		Expect(moxjson.Get(j, moxjson.ParsePath("y"))).To(BeEquivalentTo(6))
	})
})
