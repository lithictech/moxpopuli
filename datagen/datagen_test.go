package datagen_test

import (
	"context"
	"github.com/lithictech/moxpopuli/datagen"
	"github.com/lithictech/moxpopuli/fixturegen"
	"github.com/lithictech/moxpopuli/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestDatagen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "datagen Suite")
}

var _ = Describe("datagen", func() {
	ctx := context.Background()
	It("generates datas", func() {
		fixture := fixturegen.Generate(fixturegen.GenerateInput{})
		sch := schema.Derive("", fixture)
		out := datagen.Generate(ctx, datagen.GenerateInput{Schema: sch})
		Expect(out).To(HaveKeyWithValue("float", BeAssignableToTypeOf(float64(1))))
	})
})
