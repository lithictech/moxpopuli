package schemamerge_test

import (
	"context"
	"github.com/lithictech/moxpopuli/fixturegen"
	"github.com/lithictech/moxpopuli/jsonformat"
	"github.com/lithictech/moxpopuli/jsontype"
	"github.com/lithictech/moxpopuli/schema"
	"github.com/lithictech/moxpopuli/schemamerge"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestSchemamerge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "schemamerge Suite")
}

var _ = Describe("schemamerge", func() {
	ctx := context.Background()
	It("can merge full schemas", func() {
		d1 := fixturegen.Generate(fixturegen.GenerateInput{})
		d2 := fixturegen.Generate(fixturegen.GenerateInput{})
		sch1 := schema.Derive("", d1)
		sch2 := schema.Derive("", d2)
		m := schemamerge.Merge(ctx, schemamerge.MergeInput{
			Key: "",
			S1:  sch1,
			S2:  sch2,
		})
		Expect(m.Schema[schema.P_PROPERTIES]).To(And(
			HaveKeyWithValue("iso-country", HaveKeyWithValue(schema.P_MIN_LENGTH, 3)),
			HaveKeyWithValue("timestamp", HaveKeyWithValue(schema.P_TYPE, jsontype.T_INTEGER)),
		))
	})

	It("handles timezones in utc", func() {
		d1 := fixturegen.Generate(fixturegen.GenerateInput{TZ: time.UTC})
		d2 := fixturegen.Generate(fixturegen.GenerateInput{TZ: time.UTC})
		sch1 := schema.Derive("", d1)
		sch2 := schema.Derive("", d2)
		m := schemamerge.Merge(ctx, schemamerge.MergeInput{
			Key: "",
			S1:  sch1,
			S2:  sch2,
		})
		Expect(m.Schema[schema.P_PROPERTIES]).To(And(
			HaveKeyWithValue("date-time", HaveKeyWithValue(schema.P_FORMAT, jsonformat.F_DATETIME)),
			HaveKeyWithValue("date-time-notz", HaveKeyWithValue(schema.P_FORMAT, jsonformat.F_DATETIME_NOTZ)),
		))
	})
})
