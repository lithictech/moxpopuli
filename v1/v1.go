package v1

import (
	"github.com/labstack/echo"
	"github.com/lithictech/go-aperitif/api"
	"github.com/lithictech/go-aperitif/api/apiparams"
	"github.com/lithictech/moxpopuli/asyncapispecmerge"
	"github.com/lithictech/moxpopuli/datagen"
	"github.com/lithictech/moxpopuli/moxio"
	"github.com/lithictech/moxpopuli/schema"
	"github.com/lithictech/moxpopuli/schemamerge"
	"github.com/pkg/errors"
	"github.com/rgalanakis/sashay"
	"net/http"
	"strings"
)

func MountSwaggerui(e *echo.Echo) {
	e.Static("/swaggerui", "v1/swaggerui")
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasSuffix(strings.TrimSuffix(c.Request().URL.Path, "/"), "swaggerui") {
				return c.Redirect(302, "/swaggerui/index.html")
			}
			return next(c)
		}
	})
}

func NewSashay() *sashay.Sashay {
	sa := sashay.New(
		"WebhookDB Mox Populi",
		"Convert real-world events into AsyncAPI specifications. "+
			"See https://github.com/lithictech/moxpopuli for more. "+
			"Brought to you by WebhookDB, https://webhookdb.com",
		"1.0.0").
		SetContact("WebhookDB", "https://webhookdb.com", "hello@webhookdb.com").
		SetLicense("MIT", "https://github.com/lithictech/moxpopuli/blob/main/LICENSE").
		AddServer("https://moxpopuli.webhookdb.com", "Production API server")
	sa.DefineDataType(schema.Schema{}, sashay.SimpleDataTyper("object", ""))
	sa.DefineDataType(map[string]string{}, sashay.SimpleDataTyper("object", ""))
	sa.Add(quickstartSchemagenOp)
	sa.Add(schemagenOp)
	sa.Add(specgenOp)
	sa.Add(datagenOp)
	return sa
}

func Register(e *echo.Echo) {
	h := handlers{}
	e.Add(schemagenOp.Method, schemagenOp.Path, h.schemagen)
	e.Add(quickstartSchemagenOp.Method, quickstartSchemagenOp.Path, h.quickstartSchemagen)
	e.Add(specgenOp.Method, specgenOp.Path, h.specgen)
	e.Add(datagenOp.Method, datagenOp.Path, h.datagen)
}

type handlers struct{}

var quickstartSchemagenOp = sashay.NewOperation(
	"POST",
	"/v1/schemagen/quickstart",
	"Generate JSONSchema for arbitrary JSON payloads. Post an array of JSON directly. The body is the JSONSchema MoxPopuli derives for the body.",
	nil, //[]interface{}{},
	nil, //schema.Schema{},
	api.Error{},
)

func (h handlers) quickstartSchemagen(c echo.Context) error {
	ctx := api.StdContext(c)
	var params []interface{}
	if err := c.Bind(&params); err != nil {
		return err
	}
	payloadIterator := moxio.NewMemoryIterator(params)
	mergeResult, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
		PayloadIterator: payloadIterator,
	})
	if err != nil {
		return err
	}
	return c.JSONPretty(200, mergeResult.Schema, "  ")
}

var schemagenOp = sashay.NewOperation(
	"POST",
	"/v1/schemagen",
	"Incrementally generate JSONSchema for JSON payloads.",
	SchemagenParams{},
	SchemagenResponse{},
	api.Error{},
)

type SchemagenParams struct {
	Schema        schema.Schema `json:"schema" description:"The existing schema, if any. You can save the 'schema' from the response and then submit it in later requests."`
	Payloads      []interface{} `json:"payloads" description:"Array of JSON events. Mox Populi iteratively merges these into the schema."`
	ExamplesLimit *int          `json:"examples_limit" validate:"min=0,max=10" description:"How many examples to include in the resulting schema. See README for details about example sampling."`
}
type SchemagenResponse struct {
	Schema schema.Schema `json:"schema" description:"The JSONSchema derived from the input schema (if any) and each payload."`
}

func (h handlers) schemagen(c echo.Context) error {
	ctx := api.StdContext(c)
	var params SchemagenParams
	if err := apiparams.BindAndValidate(apiParamsAdapter{}, &params, c); err != nil {
		return err
	}
	payloadIterator := moxio.NewMemoryIterator(params.Payloads)
	mergeResult, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
		Schema:          params.Schema,
		PayloadIterator: payloadIterator,
		ExampleLimit:    params.ExamplesLimit,
	})
	if err != nil {
		return err
	}
	resp := SchemagenResponse{Schema: mergeResult.Schema}
	return c.JSONPretty(200, resp, "  ")
}

var specgenOp = sashay.NewOperation(
	"POST",
	"/v1/specgen",
	"Generate an AsyncAPI spec based on events for the supported protocol. Returns the new AsyncAPI spec.",
	SpecgenParams{},
	SpecgenResponse{},
	api.Error{},
)

type SpecgenParams struct {
	ExamplesLimit *int                               `json:"examples_limit" validate:"min=0|max=10" description:"See /schemagen for an explanation of this parameter."`
	Protocol      string                             `json:"protocol" enum:"http" description:"The protocol/binding to use to use when generating the spec. The value here determines which event array is used."`
	Specification map[string]interface{}             `json:"specification" description:"The existing AsyncAPI spec, if any. Generally you at least must supply the 'info' section. Everything else can usually be determined through the events."`
	HttpEvents    []asyncapispecmerge.MergeHttpEvent `json:"http_events" description:"Events to use for the 'http' protocol."`
}

type SpecgenResponse struct {
	Specification map[string]interface{} `json:"specification"`
}

func (h handlers) specgen(c echo.Context) error {
	ctx := api.StdContext(c)
	var params SpecgenParams
	if err := apiparams.BindAndValidate(apiParamsAdapter{}, &params, c); err != nil {
		return err
	}
	var events []interface{}
	var merge asyncapispecmerge.Merge
	if params.Protocol == "http" {
		merge = asyncapispecmerge.MergeHttp
		events = make([]interface{}, len(params.HttpEvents))
		for i, e := range params.HttpEvents {
			events[i] = e
		}
	} else {
		return errors.New("unsupported binding, should have been validated")
	}
	spec := params.Specification
	if spec == nil {
		spec = make(map[string]interface{}, 2)
	}
	if err := merge(ctx, asyncapispecmerge.MergeInput{
		Spec:          spec,
		EventIterator: moxio.NewMemoryIterator(events),
		ExampleLimit:  params.ExamplesLimit,
	}); err != nil {
		return errors.Wrap(err, "merging")
	}
	resp := SpecgenResponse{Specification: spec}
	return c.JSONPretty(200, resp, "  ")
}

var datagenOp = sashay.NewOperation(
	"POST",
	"/v1/datagen",
	"Generate fixtured data for a JSONSchema.",
	DatagenParams{},
	DatagenResponse{},
	api.Error{},
)

type DatagenParams struct {
	Schema schema.Schema `json:"schema"`
	Count  int           `json:"count" default:"5"`
}

type DatagenResponse struct {
	Items []interface{} `json:"items"`
}

func (h handlers) datagen(c echo.Context) error {
	ctx := api.StdContext(c)
	var params DatagenParams
	if err := apiparams.BindAndValidate(apiParamsAdapter{}, &params, c); err != nil {
		return err
	}
	items := make([]interface{}, params.Count)
	for i := 0; i < params.Count; i++ {
		items[i] = datagen.Generate(ctx, datagen.GenerateInput{Schema: params.Schema})
	}
	resp := DatagenResponse{Items: items}
	return c.JSONPretty(200, resp, "  ")
}

type apiParamsAdapter struct{}

func (apiParamsAdapter) Request(handlerArgs []interface{}) *http.Request {
	return handlerArgs[0].(echo.Context).Request()
}
func (apiParamsAdapter) RouteParamNames(handlerArgs []interface{}) []string {
	return handlerArgs[0].(echo.Context).ParamNames()
}
func (apiParamsAdapter) RouteParamValues(handlerArgs []interface{}) []string {
	return handlerArgs[0].(echo.Context).ParamValues()
}
