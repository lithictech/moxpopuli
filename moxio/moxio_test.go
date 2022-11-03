package moxio_test

import (
	"context"
	"github.com/lithictech/moxpopuli/moxio"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"os"
	"testing"
)

func TestMoxio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "moxio Suite")
}

var _ = Describe("loaders", func() {
	ctx := context.Background()
	Describe("file protocol", func() {
		It("noops if the file is not present", func() {
			ld, err := moxio.LoadOne(ctx, "file://./../testdata/noexist.json", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(ld).To(BeNil())
		})
		It("loads if the file is present", func() {
			ld, err := moxio.LoadOne(ctx, "file://./../testdata/fixturedemo.schema.json", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(ld).To(BeAssignableToTypeOf(map[string]interface{}{}))
		})
	})
})

var _ = Describe("savers", func() {
	ctx := context.Background()
	Describe("file protocol", func() {
		It("saves to the file", func() {
			tf, err := os.CreateTemp("", ".json")
			Expect(err).ToNot(HaveOccurred())
			Expect(moxio.Save(ctx, "file://"+tf.Name(), "", "contents")).To(Succeed())
			tf.Seek(0, 0)
			b, err := io.ReadAll(tf)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(b)).To(Equal(`"contents"` + "\n"))
		})
	})
})
