package schema

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/lithictech/moxpopuli/internal"
	"github.com/lithictech/moxpopuli/jsonformat"
	"github.com/lithictech/moxpopuli/jsontype"
	"github.com/lithictech/moxpopuli/moxio"
	"github.com/lithictech/moxpopuli/moxjson"
	"github.com/lithictech/moxpopuli/redact"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Field string

//goland:noinspection GoSnakeCaseUsage
const (
	P_ENUM       Field = "enum"
	P_EXAMPLES   Field = "examples"
	P_FORMAT     Field = "format"
	P_ITEMS      Field = "items"
	P_MINIMUM    Field = "minimum"
	P_MIN_LENGTH Field = "minLength"
	P_MAXIMUM    Field = "maximum"
	P_MAX_LENGTH Field = "maxLength"
	P_ONE_OF     Field = "oneOf"
	P_PROPERTIES Field = "properties"
	P_TYPE       Field = "type"

	PX_IDENTIFIER      Field = "x-identifier"
	PX_NULLABLE        Field = "x-nullable"
	PX_LAST_VALUE      Field = "x-lastValue"
	PX_SAMPLES         Field = "x-samples"
	PX_SEEN_MINIMUM    Field = "x-seenMinimum"
	PX_SEEN_MAXIMUM    Field = "x-seenMaximum"
	PX_SEEN_MIN_LENGTH Field = "x-seenMinLength"
	PX_SEEN_MAX_LENGTH Field = "x-seenMaxLength"
	PX_SEEN_STRINGS    Field = "x-seenStrings"
	PX_SENSITIVE       Field = "x-sensitive"
	PX_URI_LOCATIONS   Field = "x-uriLocations"
)

type Schema map[Field]interface{}

func FromMap(m map[string]interface{}) Schema {
	r := make(Schema, len(m))
	for k, v := range m {
		r[Field(k)] = v
	}
	return r
}

func Read(r io.Reader) (Schema, error) {
	var sch Schema
	if err := json.NewDecoder(r).Decode(&r); err != nil {
		return sch, err
	}
	return sch, nil
}

func Parse(s string) (sch Schema, err error) {
	err = json.Unmarshal([]byte(s), &sch)
	return
}

func Dump(sch Schema, w io.Writer) error {
	return moxjson.NewPrettyEncoder(w).Encode(sch)
}

func Load(ctx context.Context, uri, arg string) (Schema, error) {
	o, err := moxio.LoadOneMap(ctx, uri, arg)
	if err != nil {
		return nil, err
	}
	return FromMap(o), nil
}

func Coerce(i interface{}) Schema {
	if s, ok := i.(Schema); ok {
		return s
	}
	if m, ok := i.(map[string]interface{}); ok {
		return FromMap(m)
	}
	internal.Assert(false, "cannot coerce %v", i)
	return nil
}

func CoerceSlice(i interface{}) []Schema {
	if si, ok := i.([]Schema); ok {
		return si
	}
	ifaceSlice, ok := i.([]interface{})
	internal.Assert(ok, "invalid schema slice: %v", i)
	r := make([]Schema, len(ifaceSlice))
	for i, iface := range ifaceSlice {
		r[i] = Coerce(iface)
	}
	return r
}

func (s Schema) Type() jsontype.JsonType {
	if i, ok := s[P_TYPE]; !ok {
		return jsontype.T_NOTYPE
	} else if t, ok := i.(jsontype.JsonType); ok {
		return t
	} else {
		return jsontype.JsonType(i.(string))
	}
}

func (s Schema) Format() jsonformat.JsonFormat {
	if i, ok := s[P_FORMAT]; !ok {
		return jsonformat.F_NOFORMAT
	} else if t, ok := i.(jsonformat.JsonFormat); ok {
		return t
	} else {
		return jsonformat.JsonFormat(i.(string))
	}
}

func (s Schema) DeepClone() Schema {
	c := make(Schema, len(s))
	for k, v := range s {
		if vs, ok := v.(Schema); ok {
			c[k] = vs.DeepClone()
		} else {
			c[k] = v
		}
	}
	return c
}

func (s Schema) NullOnly() bool {
	return s.Nullable() && s.Type() == jsontype.T_NOTYPE

}

func (s Schema) Nullable() bool {
	x, ok := s[PX_NULLABLE]
	if !ok {
		return false
	}
	return x.(bool)
}

func (s Schema) Samples() int {
	if _, ok := s[PX_SAMPLES]; !ok {
		return 0
	}
	return *unwrapIntPtr(s, PX_SAMPLES)
}

func (s Schema) IncrSamples() {
	s[PX_SAMPLES] = s.Samples() + 1
}

func (s Schema) IncrSamplesDeep() {
	s.IncrSamples()
	for _, v := range s {
		if sv, ok := v.(Schema); ok {
			sv.IncrSamplesDeep()
		}
	}
}

func (s Schema) ToInteger() (IntegerSchema, bool) {
	return IntegerSchema(s), s.Type() == jsontype.T_INTEGER
}

func (s Schema) ToNumber() (NumberSchema, bool) {
	return NumberSchema(s), s.Type() == jsontype.T_NUMBER
}

func (s Schema) ToString() (StringSchema, bool) {
	return StringSchema(s), s.Type() == jsontype.T_STRING
}

func (s Schema) ToBoolean() (BooleanSchema, bool) {
	return BooleanSchema(s), s.Type() == jsontype.T_BOOLEAN
}

func (s Schema) ToObject() (ObjectSchema, bool) {
	return ObjectSchema(s), s.Type() == jsontype.T_OBJECT
}

func (s Schema) MustObject() ObjectSchema {
	o, ok := s.ToObject()
	if !ok {
		panic("must be an object, got " + s.Type())
	}
	return o
}

func (s Schema) ToArray() (ArraySchema, bool) {
	return ArraySchema(s), s.Type() == jsontype.T_ARRAY
}

type IntegerSchema Schema

func (s IntegerSchema) Minimum() *int {
	return unwrapIntPtr(Schema(s), P_MINIMUM)
}
func (s IntegerSchema) Maximum() *int {
	return unwrapIntPtr(Schema(s), P_MAXIMUM)
}
func (s IntegerSchema) Enum() []int {
	x, ok := s[P_ENUM]
	if !ok {
		return nil
	}
	return internal.SliceIToInt(x)
}
func (s IntegerSchema) SeenMinimum() *int {
	return unwrapIntPtr(Schema(s), PX_SEEN_MINIMUM)
}
func (s IntegerSchema) SeenMaximum() *int {
	return unwrapIntPtr(Schema(s), PX_SEEN_MAXIMUM)
}

type NumberSchema Schema

func (s NumberSchema) Minimum() *float64 {
	x, ok := s[P_MINIMUM].(float64)
	return maybeFloat(x, ok)
}
func (s NumberSchema) Maximum() *float64 {
	x, ok := s[P_MAXIMUM].(float64)
	return maybeFloat(x, ok)
}
func (s NumberSchema) SeenMinimum() *float64 {
	x, ok := s[PX_SEEN_MINIMUM].(float64)
	return maybeFloat(x, ok)
}
func (s NumberSchema) SeenMaximum() *float64 {
	x, ok := s[PX_SEEN_MAXIMUM].(float64)
	return maybeFloat(x, ok)
}

type StringSchema Schema

func (s StringSchema) MinLength() *int {
	return unwrapIntPtr(Schema(s), P_MIN_LENGTH)
}
func (s StringSchema) MaxLength() *int {
	return unwrapIntPtr(Schema(s), P_MAX_LENGTH)
}
func (s StringSchema) Enum() []string {
	e, ok := s[P_ENUM]
	if !ok {
		return nil
	}
	es, ok := e.([]string)
	if ok {
		return es
	}
	return internal.SliceIToStr(e)
}

func (s StringSchema) SeenStrings() []string {
	e, ok := s[PX_SEEN_STRINGS]
	if !ok {
		return nil
	}
	es, ok := e.([]string)
	if ok {
		return es
	}
	return internal.SliceIToStr(e)
}

func (s StringSchema) SeenMinLength() *int {
	return unwrapIntPtr(Schema(s), PX_SEEN_MIN_LENGTH)
}
func (s StringSchema) SeenMaxLength() *int {
	return unwrapIntPtr(Schema(s), PX_SEEN_MAX_LENGTH)
}
func (s StringSchema) SeenMinimum() *string {
	x, ok := s[PX_SEEN_MINIMUM].(string)
	return maybeString(x, ok)
}
func (s StringSchema) SeenMaximum() *string {
	x, ok := s[PX_SEEN_MAXIMUM].(string)
	return maybeString(x, ok)
}
func (s StringSchema) SeenUriLocations() []string {
	e, ok := s[PX_URI_LOCATIONS]
	if !ok {
		return nil
	}
	return internal.SliceIToStr(e)
}

func (s StringSchema) Sensitive() bool {
	x, ok := s[PX_SENSITIVE].(bool)
	if !ok {
		return false
	}
	return x
}

type BooleanSchema Schema

type ObjectSchema Schema

func (s ObjectSchema) Properties() map[string]Schema {
	x, ok := s[P_PROPERTIES].(map[string]Schema)
	if ok {
		return x
	}
	m := s[P_PROPERTIES].(map[string]interface{})
	r := make(map[string]Schema, len(m))
	for k, v := range m {
		r[k] = FromMap(v.(map[string]interface{}))
	}
	return r
}

type ArraySchema Schema

func (s ArraySchema) Items() Schema {
	x, ok := s[P_ITEMS].(Schema)
	if ok {
		return x
	}
	return FromMap(s[P_ITEMS].(map[string]interface{}))
}
func (s ArraySchema) SeenMinLength() *int {
	return unwrapIntPtr(Schema(s), PX_SEEN_MIN_LENGTH)
}
func (s ArraySchema) SeenMaxLength() *int {
	return unwrapIntPtr(Schema(s), PX_SEEN_MAX_LENGTH)
}

type UntypedSchema Schema

func unwrapIntPtr(sch Schema, f Field) *int {
	x, ok := sch[f]
	if !ok {
		return nil
	}
	xi, ok := x.(int)
	if ok {
		return &xi
	}
	xf, ok := x.(float64)
	if ok {
		xfi := int(xf)
		return &xfi
	}
	return nil
}

func maybeFloat(i float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	return &i
}

func maybeString(i string, ok bool) *string {
	if !ok {
		return nil
	}
	return &i
}

func Derive(key string, o interface{}) Schema {
	if o == nil {
		return Schema{PX_NULLABLE: true}
	}
	o = internal.CoerceToLikelyGoType(o)
	t := jsontype.Sniff(o)
	switch t {
	case jsontype.T_BOOLEAN:
		return Schema{P_TYPE: jsontype.T_BOOLEAN}
	case jsontype.T_NUMBER:
		return deriveNumber(o.(float64))
	case jsontype.T_INTEGER:
		return deriveInteger(o.(int))
	case jsontype.T_STRING:
		return deriveString(key, o.(string))
	case jsontype.T_OBJECT:
		return deriveObject(o.(map[string]interface{}))
	case jsontype.T_ARRAY:
		return deriveArray(key, o.([]interface{}))
	default:
		// Since we are deriving here, we should never run into a notype
		panic("unhandled type " + t)
	}
}

func deriveNumber(v float64) Schema {
	f := jsonformat.Sniff(jsontype.T_NUMBER, v)
	s := Schema{
		P_TYPE:   jsontype.T_NUMBER,
		P_FORMAT: f,
	}
	s[PX_SEEN_MINIMUM] = v
	s[PX_SEEN_MAXIMUM] = v
	return s
}

func deriveInteger(v int) Schema {
	f := jsonformat.Sniff(jsontype.T_INTEGER, v)
	s := Schema{
		P_TYPE:   jsontype.T_INTEGER,
		P_FORMAT: f,
	}
	// Record these in all cases, in case we need to coerce ZERO_ONE to a number.
	s[PX_SEEN_MINIMUM] = v
	s[PX_SEEN_MAXIMUM] = v
	return s
}

func deriveString(k, v string) Schema {
	f := jsonformat.Sniff(jsontype.T_STRING, v)
	s := Schema{P_TYPE: jsontype.T_STRING}
	if sens, ok := sensitive(f, k, v); ok {
		// If our string is sensitive, do NOT analyze sensitive content.
		// Get rid of it ASAP.
		v = sens
		s[PX_SENSITIVE] = true
	}
	// Record the value so we can handle it when merging
	s[PX_SEEN_STRINGS] = []string{v}
	if f != jsonformat.F_NOFORMAT {
		s[P_FORMAT] = f
	}
	if f == jsonformat.F_COUNTRY {
		s[P_MIN_LENGTH] = 3
		s[P_MAX_LENGTH] = 3
	} else if f == jsonformat.F_CURRENCY {
		s[P_MIN_LENGTH] = 2
		s[P_MAX_LENGTH] = 2
	} else if f == jsonformat.F_UUID4 {
		s[P_MIN_LENGTH] = len(v)
		s[P_MAX_LENGTH] = len(v)
	} else if f == jsonformat.F_NUMERICAL {
		s[PX_SEEN_MINIMUM] = v
		s[PX_SEEN_MAXIMUM] = v
	} else if jsonformat.IsChronolike(f) {
		s[PX_SEEN_MINIMUM] = v
		s[PX_SEEN_MAXIMUM] = v
	} else if f == jsonformat.F_URI && !strings.HasPrefix(v, "/") {
		u, _ := url.Parse(v)
		s[PX_URI_LOCATIONS] = []interface{}{fmt.Sprintf("%s://%s", u.Scheme, u.Host)}
		// Normally we don't care about this, but we may need to merge into a string.
		s[PX_SEEN_MIN_LENGTH] = len(v)
		s[PX_SEEN_MAX_LENGTH] = len(v)
	} else {
		s[PX_SEEN_MIN_LENGTH] = len(v)
		s[PX_SEEN_MAX_LENGTH] = len(v)
	}
	return s
}

func sensitive(f jsonformat.JsonFormat, k, v string) (string, bool) {
	if jsonformat.IsChronolike(f) {
		return "", false
	}
	canonical := canonicalKey(k)
	shouldRedact := strings.HasSuffix(canonical, "token") ||
		strings.HasSuffix(canonical, "code") ||
		strings.HasSuffix(canonical, "secret") ||
		strings.HasSuffix(canonical, "digest")
	if shouldRedact {
		return redact.Zero(v), true
	}
	if len(v) < 8 {
		// Assume nothing secure is 8 chars or less
		return "", false
	}
	if redact.IsSha(v) {
		return "", false
	}
	if u, err := url.Parse(v); err == nil && u.Scheme != "" && u.Host != "" {
		if u.User == nil {
			return "", false
		}
		u.User = url.User("*")
		return u.String(), true
	}
	shouldRandomize := redact.IsGibberish(v)
	if shouldRandomize {
		return redact.UnsafeVariableHash([]byte(v), []byte(SensitiveSalt)), true
	}
	return "", false
}

var SensitiveSalt string

func init() {
	if s, ok := os.LookupEnv("MOXPOPULI_SALT"); ok {
		SensitiveSalt = s
	} else {
		h := make([]byte, 32)
		if _, err := rand.Read(h); err != nil {
			panic(err)
		}
		SensitiveSalt = base64.StdEncoding.EncodeToString(h)
	}
}

func canonicalKey(s string) string {
	return strings.ToLower(canonicalReplacement.ReplaceAllString(s, ""))
}

var canonicalReplacement = regexp.MustCompile("[^a-zA-Z0-9]+")

func deriveObject(v map[string]interface{}) Schema {
	s := Schema{
		P_TYPE: jsontype.T_OBJECT,
	}
	properties := make(map[string]Schema, len(v))
	for key, value := range v {
		properties[key] = Derive(key, value)
	}
	s[P_PROPERTIES] = properties
	return s
}

func deriveArray(k string, v []interface{}) Schema {
	s := Schema{
		P_TYPE: jsontype.T_ARRAY,
	}
	if len(v) == 0 {
		s[P_ITEMS] = Schema{}
	} else {
		s[P_ITEMS] = Derive(k, v[0])
	}
	s[PX_SEEN_MIN_LENGTH] = len(v)
	s[PX_SEEN_MAX_LENGTH] = len(v)
	return s
}

//goland:noinspection GoSnakeCaseUsage
const (
	TF_DATE          = "2006-01-02"
	TF_DATETIME      = time.RFC3339
	TF_DATETIME_NOTZ = "2006-01-02T15:04:05"
	TF_TIME          = "15:04:05Z07:00"
)

func Examples(sch Schema) []interface{} {
	i, ok := sch[P_EXAMPLES]
	if !ok {
		return nil
	}
	return i.([]interface{})
}
