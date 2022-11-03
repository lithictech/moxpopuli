package fixturegen_test

import (
	"github.com/lithictech/moxpopuli/fixturegen"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestFixturegen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fixturegen Suite")
}

var _ = Describe("fixturegen", func() {
	It("generates fixtures", func() {
		out := fixturegen.Run(fixturegen.RunInput{Count: 2})
		Expect(out).To(HaveLen(2))
		Expect(out[0]).To(HaveKeyWithValue("float", BeAssignableToTypeOf(float64(1))))
	})
	It("returns normal JSON types", func() {
		f := fixturegen.Generate(fixturegen.GenerateInput{})
		Expect(f["int32"]).To(BeAssignableToTypeOf(float64(0)))
	})
})
