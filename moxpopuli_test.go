package moxpopuli

import (
	"context"
	"fmt"
	"github.com/lithictech/moxpopuli/moxio"
	"github.com/lithictech/moxpopuli/schema"
	"github.com/lithictech/moxpopuli/schemamerge"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strconv"
	"strings"
	"testing"
)

func TestMoxpopuli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "moxpopuli Suite")
}

var _ = Describe("schemagen", func() {
	ctx := context.Background()
	dump := func(sch schema.Schema) string {
		w := strings.Builder{}
		Expect(schema.Dump(sch, &w)).To(Succeed())
		return w.String()
	}
	It("merge a payload into a fresh schema", func() {
		iter, err := moxio.LoadIterator(ctx, "_", `{"x":1}`)
		Expect(err).ToNot(HaveOccurred())
		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
			Schema:          schema.Schema{},
			PayloadIterator: iter,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"x": {
					"format": "int32",
					"type": "integer",
					"x-seenMaximum": 1,
					"x-seenMinimum": 1
				}
			},
			"type": "object",
			"x-samples": 1
		}`))
	})
	It("can merge integers into floats", func() {
		iter, err := moxio.LoadIterator(ctx, "_", `{"y":10}`+"\n"+`{"y":10.5}`)
		Expect(err).ToNot(HaveOccurred())
		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
			Schema:          schema.Schema{},
			PayloadIterator: iter,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"y": {
					"format": "float",
					"type": "number",
					"x-samples": 2,
					"x-seenMaximum": 10.5,
					"x-seenMinimum": 10
				}
			},
			"type": "object",
			"x-samples": 2
		}`))
	})
	It("can merge floats into integers", func() {
		iter, err := moxio.LoadIterator(ctx, "_", `{"y":10.1}`+"\n"+`{"y":0}`)
		Expect(err).ToNot(HaveOccurred())
		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
			Schema:          schema.Schema{},
			PayloadIterator: iter,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"y": {
					"format": "float",
					"type": "number",
					"x-samples": 2,
					"x-seenMaximum": 10.1,
					"x-seenMinimum": 0
				}
			},
			"type": "object",
			"x-samples": 2
		}`))
	})
	It("can merge urls into strings", func() {
		iter, err := moxio.LoadIterator(ctx, "_", `{"y":"https://x.y.z"}`+"\n"+`{"y":"a"}`)
		Expect(err).ToNot(HaveOccurred())
		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
			Schema:          schema.Schema{},
			PayloadIterator: iter,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"y": {
					"type": "string",
					"x-samples": 2,
					"x-seenMaxLength": 13,
					"x-seenMinLength": 1
				}
			},
			"type": "object",
			"x-samples": 2
		}`))
	})
	It("randomizes sensitive strings", func() {
		schema.SensitiveSalt = "abcd"
		iter, err := moxio.LoadIterator(ctx, "_", `{"x":"c2c691f00d678abf6c54b18fd930"}`)
		Expect(err).ToNot(HaveOccurred())
		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
			Schema:          schema.Schema{},
			PayloadIterator: iter,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"x": {
					"type": "string",
					"x-seenMaxLength": 28,
					"x-seenMinLength": 28,
					"x-seenStrings": [
						"3CSutK3KOFEaqW2zub394P5fjGDT"
					],
					"x-sensitive": true
				}
			},
			"type": "object",
			"x-samples": 1
		}`))
	})
	It("guesses at enums", func() {
		arg := strings.Builder{}
		for i := 10; i < 60; i++ {
			arg.WriteString(fmt.Sprintf(`{"x":"VALUE_%s"}%s`, string(strconv.Itoa(i)[0]), "\n"))
		}
		iter, err := moxio.LoadIterator(ctx, "_", arg.String())
		Expect(err).ToNot(HaveOccurred())
		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
			Schema:          schema.Schema{},
			PayloadIterator: iter,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"x": {
					"enum": [
						"VALUE_1",
						"VALUE_2",
						"VALUE_3",
						"VALUE_4",
						"VALUE_5"
					],
					"type": "string",
					"x-samples": 50,
					"x-seenMaxLength": 7,
					"x-seenMinLength": 7
				}
			},
			"type": "object",
			"x-samples": 50
		}`))
	})
	It("consistently hashes within a single process", func() {
		schema.SensitiveSalt = "abcd"
		schout1 := schema.Derive("", map[string]interface{}{"x": "c2c691f00d678abf6c54b18fd930"})
		schout2 := schema.Derive("", map[string]interface{}{"x": "c2c691f00d678abf6c54b18fd930"})
		Expect(dump(schout1)).To(MatchJSON(`{
			"properties": {
				"x": {
					"type": "string",
					"x-seenMaxLength": 28,
					"x-seenMinLength": 28,
					"x-seenStrings": [
						"3CSutK3KOFEaqW2zub394P5fjGDT"
					],
					"x-sensitive": true
				}
			},
			"type": "object"
		}`))
		Expect(dump(schout2)).To(MatchJSON(`{
			"properties": {
				"x": {
					"type": "string",
					"x-seenMaxLength": 28,
					"x-seenMinLength": 28,
					"x-seenStrings": [
						"3CSutK3KOFEaqW2zub394P5fjGDT"
					],
					"x-sensitive": true
				}
			},
			"type": "object"
		}`))
		schema.SensitiveSalt = "xyz"
		schout3 := schema.Derive("", map[string]interface{}{"x": "c2c691f00d678abf6c54b18fd930"})
		Expect(dump(schout3)).To(MatchJSON(`{
			"properties": {
				"x": {
					"type": "string",
					"x-seenMaxLength": 28,
					"x-seenMinLength": 28,
					"x-seenStrings": [
						"7IWIqRa7Q6q2JiImGD3VQ3Guxa2a"
					],
					"x-sensitive": true
				}
			},
			"type": "object"
		}`))
	})
	It("assumes zero-one at some point", func() {
		lines := strings.Join([]string{`{"x":0}`, `{"x":0}`}, "\n")
		iter, err := moxio.LoadIterator(ctx, "_", lines)
		Expect(err).ToNot(HaveOccurred())
		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{Schema: schema.Schema{}, PayloadIterator: iter})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"x": {
					"format": "int32",
					"type": "integer",
					"x-samples": 2,
					"x-seenMaximum": 0,
					"x-seenMinimum": 0
				}
			},
			"type": "object",
			"x-samples": 2
		}`))

		iter, err = moxio.LoadIterator(ctx, "_", strings.Join([]string{`{"x":1}`}, "\n"))
		Expect(err).ToNot(HaveOccurred())
		schout, err = schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{Schema: schout.Schema, PayloadIterator: iter})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"x": {
					"format": "int32",
					"type": "integer",
					"x-samples": 3,
					"x-seenMaximum": 1,
					"x-seenMinimum": 0
				}
        },
        "type": "object",
        "x-samples": 3
		}`))

		iter, err = moxio.LoadIterator(ctx, "_", strings.Join([]string{`{"x":0}`, `{"x":1}`, `{"x":0}`}, "\n"))
		Expect(err).ToNot(HaveOccurred())
		schout, err = schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{Schema: schout.Schema, PayloadIterator: iter})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"x": {
					"enum": [
						0,
						1
					],
					"format": "zero-one",
					"type": "integer",
					"x-samples": 6,
					"x-seenMaximum": 1,
					"x-seenMinimum": 0
				}
			},
			"type": "object",
			"x-samples": 6
		}`))

		iter, err = moxio.LoadIterator(ctx, "_", strings.Join([]string{`{"x":2}`}, "\n"))
		Expect(err).ToNot(HaveOccurred())
		schout, err = schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{Schema: schout.Schema, PayloadIterator: iter})
		Expect(err).ToNot(HaveOccurred())
		Expect(dump(schout.Schema)).To(MatchJSON(`{
			"properties": {
				"x": {
					"format": "int32",
					"type": "integer",
					"x-samples": 7,
					"x-seenMaximum": 2,
					"x-seenMinimum": 0
				}
			},
			"type": "object",
			"x-samples": 7
		}`))
	})
})
