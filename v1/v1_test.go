package v1_test

import (
	"github.com/labstack/echo"
	"github.com/lithictech/go-aperitif/api"
	. "github.com/lithictech/go-aperitif/api/echoapitest"
	. "github.com/lithictech/go-aperitif/apitest"
	v1 "github.com/lithictech/moxpopuli/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rgalanakis/golangal"
	"github.com/rgalanakis/sashay"
	"testing"
)

func TestV1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "v1 Suite")
}

type anymap map[string]interface{}

var _ = Describe("v1", func() {
	var e *echo.Echo
	BeforeEach(func() {
		e = api.New(api.Config{})
		v1.Register(e)
	})
	It("can create a new Sashay", func() {
		Expect(v1.NewSashay()).To(BeAssignableToTypeOf(&sashay.Sashay{}))
	})
	Describe("POST /v1/schemagen", func() {
		It("generates schemas with no schema input", func() {
			req := NewRequest("POST", "/v1/schemagen", MustMarshal(anymap{
				"payloads": []anymap{{"x": 1}},
			}), JsonReq())
			rr := Serve(e, req)
			Expect(rr).To(HaveResponseCode(200))
			Expect(rr.Body.String()).To(MatchJSON(`{
	"schema": {
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
	}
}`))
		})
		It("generates schema from an existing schema", func() {
			req := NewRequest("POST", "/v1/schemagen", MustMarshal(anymap{
				"schema":         anymap{"type": "object", "properties": anymap{"original": anymap{"type": "boolean"}}},
				"payloads":       []anymap{{"x": 1}},
				"examples_limit": 1,
			}), JsonReq())
			rr := Serve(e, req)
			Expect(rr).To(HaveResponseCode(200))
			Expect(rr.Body.String()).To(MatchJSON(`{
	"schema": {
		"examples": [],
		"properties": {
			"original": {
				"type": "boolean"
			},
			"x": {
				"format": "int32",
				"type": "integer",
				"x-seenMaximum": 1,
				"x-seenMinimum": 1
			}
		},
		"type": "object",
		"x-samples": 2
	}
}`))
		})
	})
	Describe("POST /v1/schemagen/quickstart", func() {
		It("generates unnested schema output from array input", func() {
			req := NewRequest("POST", "/v1/schemagen/quickstart", MustMarshal([]anymap{{"x": 1}}), JsonReq())
			rr := Serve(e, req)
			Expect(rr).To(HaveResponseCode(200))
			Expect(rr.Body.String()).To(MatchJSON(`{
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
	})
	Describe("POST /v1/specgen", func() {
		It("generates specs with no spec input", func() {
			req := NewRequest("POST", "/v1/specgen", MustMarshal(anymap{
				"protocol": "http",
				"http_events": []anymap{
					{
						"method":  "PATCH",
						"path":    "/myapi",
						"headers": anymap{"X-My-Api": "1", "X-Trace-Id": "abc"},
						"payload": anymap{"x": true},
					},
				},
			}), JsonReq())
			rr := Serve(e, req)
			Expect(rr).To(HaveResponseCode(200))
			Expect(rr.Body.String()).To(ContainSubstring(`"location": "$message.header#/X-Trace-Id"`))
		})
		It("generates schema from an existing schema", func() {
			req := NewRequest("POST", "/v1/specgen", MustMarshal(anymap{
				"protocol": "http",
				"http_events": []anymap{
					{
						"method":  "PATCH",
						"path":    "/myapi",
						"headers": anymap{"X-My-Api": "1", "X-Trace-Id": "abc"},
						"payload": anymap{"x": true},
					},
				},
				"specification": anymap{"info": anymap{"title": "here is my title"}},
			}), JsonReq())
			rr := Serve(e, req)
			Expect(rr).To(HaveResponseCode(200))
			Expect(rr.Body.String()).To(ContainSubstring(`"location": "$message.header#/X-Trace-Id"`))
			Expect(rr.Body.String()).To(ContainSubstring(`"title": "here is my title"`))
		})
	})
	Describe("POST /v1/datagen", func() {
		It("generates fixtured data", func() {
			req := NewRequest("POST", "/v1/datagen", MustMarshal(anymap{
				"schema": MustUnmarshal(`{
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
	}`),
			}), JsonReq())
			rr := Serve(e, req)
			Expect(rr).To(HaveResponseCode(200))
			Expect(rr.Body.String()).To(MatchJSON(`{
        "items": [
          {
            "x": 1
          },
          {
            "x": 1
          },
          {
            "x": 1
          },
          {
            "x": 1
          },
          {
            "x": 1
          }
        ]
      }`))
		})
		It("can use an empty schema", func() {
			req := NewRequest("POST", "/v1/datagen", MustMarshal(anymap{
				"count": 2,
			}), JsonReq())
			rr := Serve(e, req)
			Expect(rr).To(HaveResponseCode(200))
			Expect(rr.Body.String()).To(MatchJSON(`{
        "items": [
          null,
          null
        ]
      }`))
		})
	})
})
