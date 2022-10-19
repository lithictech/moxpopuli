package schemamerge

import (
	"context"
	"fmt"
	"github.com/lithictech/moxpopuli/fp"
	"github.com/lithictech/moxpopuli/internal"
	. "github.com/lithictech/moxpopuli/jsonformat"
	"github.com/lithictech/moxpopuli/jsontype"
	"github.com/lithictech/moxpopuli/moxio"
	. "github.com/lithictech/moxpopuli/schema"
	"github.com/lithictech/moxpopuli/timestring"
	"github.com/pkg/errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type MergeInput struct {
	// If these schemas are an object's values, this is the property name.
	Key string
	// One of the schemas, usually the "source" (existing schema).
	S1 Schema
	// The other schema, usually the newly derived one.
	S2 Schema
}

type MergeOutput struct {
	Schema      Schema
	TypeChanged bool
}

func Merge(ctx context.Context, in MergeInput) MergeOutput {
	sr := Schema{}
	s1 := in.S1
	s2 := in.S2
	// Schemas are only empty initially.
	// Even typeless objects have custom properties,
	// like nullable or seen types.
	s1Empty := len(s1) == 0
	internal.Assert(len(s2) > 0, "the second schema should never be empty")
	if s1Empty {
		// Make sure all the subschemas get incremented,
		// since it's the first time we've seen them.
		sr = s2.DeepClone()
		sr.IncrSamplesDeep()
		return MergeOutput{Schema: sr, TypeChanged: true}
	}
	// If one or the other are 'null only' because there was no value,
	// return the other one.
	if s1.NullOnly() && !s2.NullOnly() {
		s2 = s2.DeepClone()
		s2[PX_NULLABLE] = true
		s2.IncrSamples()
		return MergeOutput{Schema: s2, TypeChanged: true}
	} else if !s1.NullOnly() && s2.NullOnly() {
		s1 = s1.DeepClone()
		s1[PX_NULLABLE] = true
		s1.IncrSamples()
		return MergeOutput{Schema: s1, TypeChanged: true}
	}
	mo := MergeOutput{Schema: sr}
	var convertIntToFloat IntegerSchema
	if s1.Type() == jsontype.T_INTEGER && s2.Type() == jsontype.T_NUMBER {
		convertIntToFloat, _ = s1.ToInteger()
	} else if s1.Type() == jsontype.T_NUMBER && s2.Type() == jsontype.T_INTEGER {
		convertIntToFloat, _ = s2.ToInteger()
	} else if s1.Type() != s2.Type() {
		// If types are not the same, compose the schemas with oneOf.
		// Assume s1 here already has samples, since it was originally added
		// as a homogenous schema.
		// We must record the samples on the new schema though,
		// since it's the first time we're seeing it.
		// NOTE: I'm not certain this is right, it may need to be in mergeSliceProperty.
		s2 = s2.DeepClone()
		s2.IncrSamples()
		sr[P_ONE_OF], mo.TypeChanged = mergeSliceProperty(ctx, P_ONE_OF, s1, s2)
		return mo
	}
	if convertIntToFloat != nil {
		convertIntToFloat[P_TYPE] = jsontype.T_NUMBER
		setIfNotNull(Schema(convertIntToFloat), P_MINIMUM, internal.IntPtrToFloatPtr(convertIntToFloat.Minimum()))
		setIfNotNull(Schema(convertIntToFloat), P_MAXIMUM, internal.IntPtrToFloatPtr(convertIntToFloat.Maximum()))
		setIfNotNull(Schema(convertIntToFloat), PX_SEEN_MINIMUM, internal.IntPtrToFloatPtr(convertIntToFloat.SeenMinimum()))
		setIfNotNull(Schema(convertIntToFloat), PX_SEEN_MAXIMUM, internal.IntPtrToFloatPtr(convertIntToFloat.SeenMaximum()))
	}
	// s1 may or may not have samples, depending on how it came in
	// (ie load, subschema merge, etc), so default to 1
	sr[PX_SAMPLES] = internal.MaxInt(s1.Samples(), 1)
	sr.IncrSamples()

	if s1.Nullable() || s2.Nullable() {
		sr[PX_NULLABLE] = true
	}
	if t := s1.Type(); t != jsontype.T_NOTYPE {
		// Generally this means the schemas are both from nulls
		sr[P_TYPE] = t
	}
	jfmt := MergeFormat(s1.Format(), s2.Format())
	if jfmt != F_NOFORMAT {
		sr[P_FORMAT] = jfmt
	}
	if s2t, ok := s2.ToInteger(); ok {
		if jfmt == F_ZERO_ONE {
			sr[P_ENUM] = s2t.Enum()
		} else {
			s1t, _ := s1.ToInteger()
			setIfNotNull(sr, P_MINIMUM, internal.MinIntPtr(s1t.Minimum(), s2t.Minimum()))
			setIfNotNull(sr, P_MAXIMUM, internal.MaxIntPtr(s1t.Maximum(), s2t.Maximum()))
			setIfNotNull(sr, PX_SEEN_MINIMUM, internal.MinIntPtr(s1t.SeenMinimum(), s2t.SeenMinimum()))
			setIfNotNull(sr, PX_SEEN_MAXIMUM, internal.MaxIntPtr(s1t.SeenMaximum(), s2t.SeenMaximum()))
		}
	} else if s2t, ok := s2.ToNumber(); ok {
		s1t, _ := s1.ToNumber()
		setIfNotNull(sr, P_MINIMUM, internal.MinFloat64Ptr(s1t.Minimum(), s2t.Minimum()))
		setIfNotNull(sr, P_MAXIMUM, internal.MaxFloat64Ptr(s1t.Maximum(), s2t.Maximum()))
		setIfNotNull(sr, PX_SEEN_MINIMUM, internal.MinFloat64Ptr(s1t.SeenMinimum(), s2t.SeenMinimum()))
		setIfNotNull(sr, PX_SEEN_MAXIMUM, internal.MaxFloat64Ptr(s1t.SeenMaximum(), s2t.SeenMaximum()))
	} else if s2t, ok := s2.ToString(); ok {
		s1t, _ := s1.ToString()
		if jfmt != F_URI {
			delete(sr, PX_URI_LOCATIONS)
		} else if uriLocs := internal.UniqueSortedStrings(append(s1t.SeenUriLocations(), s2t.SeenUriLocations()...)); len(uriLocs) > 0 {
			sr[PX_URI_LOCATIONS] = uriLocs
		}
		if jfmt == F_NUMERICAL {
			handleNumerical(sr, s1t, s2t)
		} else if timecomp, ok := timeFormatValueComparers[jfmt]; ok {
			sr[PX_SEEN_MINIMUM], _ = timecomp(internal.CompactStrings(s1t.SeenMinimum(), s2t.SeenMinimum())...)
			_, sr[PX_SEEN_MAXIMUM] = timecomp(internal.CompactStrings(s1t.SeenMaximum(), s2t.SeenMaximum())...)
		}
		setIfNotNull(sr, P_MIN_LENGTH, internal.MinIntPtr(s1t.MinLength(), s2t.MinLength()))
		setIfNotNull(sr, P_MAX_LENGTH, internal.MaxIntPtr(s1t.MaxLength(), s2t.MaxLength()))
		setIfNotNull(sr, PX_SEEN_MIN_LENGTH, internal.MinIntPtr(s1t.SeenMinLength(), s2t.SeenMinLength()))
		setIfNotNull(sr, PX_SEEN_MAX_LENGTH, internal.MaxIntPtr(s1t.SeenMaxLength(), s2t.SeenMaxLength()))
		if s1t.Sensitive() || s2t.Sensitive() {
			sr[PX_SENSITIVE] = true
		}
		if enum := append(s1t.Enum(), s2t.Enum()...); len(enum) > 0 {
			sr[P_ENUM] = enum
		}
		if seen := append(s1t.SeenStrings(), s2t.SeenStrings()...); len(seen) > 0 {
			sr[PX_SEEN_STRINGS] = seen
		}
	} else if s2t, ok := s2.ToObject(); ok {
		s1t, _ := s1.ToObject()
		sr[P_PROPERTIES], mo.TypeChanged = mergeObjects(ctx, s1t, s2t)
	} else if s2t, ok := s2.ToArray(); ok {
		s1t, _ := s1.ToArray()
		// We can end up with an empty 'items' schema if the payload that
		// the schema was derived from was empty.
		s1items, s2items := s1t.Items(), s2t.Items()
		s1itemlen, s2itemlen := len(s1items), len(s2items)
		if s1itemlen == 0 && s2itemlen == 0 {
			// If both are empty, continue an empty schema
			sr[P_ITEMS] = s1items
		} else if s1itemlen == 0 {
			// otherwise use the first present schema, and record a change
			sr[P_ITEMS] = s2t.Items()
			mo.TypeChanged = true
		} else if s2itemlen == 0 {
			// s2 is empty, use s1 items
			sr[P_ITEMS] = s1t.Items()
			mo.TypeChanged = true
		} else {
			// Both have a schema, merge it and record if the type changed
			amo := Merge(ctx, MergeInput{Key: string(P_ITEMS), S1: s1t.Items(), S2: s2t.Items()})
			sr[P_ITEMS] = amo.Schema
			mo.TypeChanged = amo.TypeChanged
		}
		setIfNotNull(sr, PX_SEEN_MIN_LENGTH, internal.MinIntPtr(s1t.SeenMinLength(), s2t.SeenMinLength()))
		setIfNotNull(sr, PX_SEEN_MAX_LENGTH, internal.MaxIntPtr(s1t.SeenMaxLength(), s2t.SeenMaxLength()))
	}
	postprocess(in.Key, sr)
	return mo
}

func handleNumerical(sr Schema, t1 StringSchema, t2 StringSchema) {
	t1min, _ := strconv.Atoi(*t1.SeenMinimum())
	t1max, _ := strconv.Atoi(*t1.SeenMaximum())
	t2min, _ := strconv.Atoi(*t2.SeenMinimum())
	t2max, _ := strconv.Atoi(*t2.SeenMaximum())
	sr[PX_SEEN_MINIMUM] = strconv.Itoa(internal.MinInt(t1min, t2min))
	sr[PX_SEEN_MAXIMUM] = strconv.Itoa(internal.MaxInt(t1max, t2max))
}

func mergeObjects(ctx context.Context, s1, s2 ObjectSchema) (map[string]Schema, bool) {
	s1props := s1.Properties()
	s2props := s2.Properties()
	result := make(map[string]Schema, len(s1props))
	typeChanged := false
	for k, s1v := range s1props {
		if s2v, ok := s2props[k]; ok {
			mo := Merge(ctx, MergeInput{Key: k, S1: s1v, S2: s2v})
			result[k] = mo.Schema
			typeChanged = typeChanged || mo.TypeChanged
		} else {
			result[k] = s1v
		}
	}
	for k, s2v := range s2props {
		if _, ok := s1props[k]; !ok {
			result[k] = s2v
		}
	}
	return result, typeChanged
}

func setIfNotNull[T *int | *float64](sch Schema, f Field, i T) {
	if i == nil {
		return
	}
	sch[f] = reflect.ValueOf(i).Elem().Interface()
}

func postprocess(key string, s Schema) {
	handleIdentifier(key, s)
	handleStringEnum(key, s)
	handleZeroOne(key, s)
}

func handleIdentifier(key string, s Schema) {
	if strings.HasSuffix(key, "_id") {
		s[PX_IDENTIFIER] = true
		return
	}
	if s.Samples() < 5 {
		return
	}
	ss, ok := s.ToString()
	if !ok {
		return
	}
	minlen, maxlen := ss.SeenMinLength(), ss.SeenMaxLength()
	if minlen == nil || maxlen == nil {
		return
	}
	if *minlen != *maxlen {
		return
	}
	if *minlen > 8 {
		s[PX_IDENTIFIER] = true
	}
}

func handleStringEnum(_ string, sch Schema) {
	ssch, ok := sch.ToString()
	if !ok {
		// Only handle strings right now
		return
	}
	// If this is a sensitive string or an identifier, we know it cannot be an enum.
	if sch.Format() == F_UUID4 || ssch.Sensitive() {
		delete(sch, P_ENUM)
		delete(sch, PX_SEEN_STRINGS)
		return
	}
	// We have enums in two fields: 'enum', and 'x-seenStrings'
	// We write x-seenStrings as we derive the schemas;
	// then when we merge, we first combine it with enum, and then:
	// - replace enum and remove seenStrings if we think we have good enums
	// - delete enum and seenStrings if we know we don't have enums
	// - delete enums and write seenStrings if we aren't sure.
	allEnums := internal.UniqueSortedStrings(append(ssch.Enum(), ssch.SeenStrings()...))
	allLikely := true
	for _, s := range allEnums {
		if !validEnumRegex.MatchString(s) {
			delete(sch, P_ENUM)
			delete(sch, PX_SEEN_STRINGS)
			return
		}
		allLikely = allLikely && likelyEnumRegex.MatchString(s)
	}
	samples := sch.Samples()
	if samples <= 10 {
		// Maybe this is an enum, we'll know more later when we've sampled more.
		delete(sch, P_ENUM)
		sch[PX_SEEN_STRINGS] = allEnums
		return
	}
	// If we every value is likely an enum, and we don't have *too* many,
	// let's treat it as an enum.
	if allLikely && len(allEnums) < 20 {
		delete(sch, PX_SEEN_STRINGS)
		sch[P_ENUM] = allEnums
		return
	}
	// Otherwise, keep accumulating seen strings until we can make a decision.
	delete(sch, P_ENUM)
	sch[PX_SEEN_STRINGS] = allEnums
	return
}

// valid enums start with a letter and contain only upper OR lowercase,
// plus numbers and underscores
var validEnumRegex = regexp.MustCompile("^[a-zA-Z]([a-z0-9_]|[A-Z0-9_])+$")

// Likely enums are short and letters-only.
var likelyEnumRegex = regexp.MustCompile("^([a-z\\d_]|[A-Z\\d_]){2,26}$")

func handleZeroOne(_ string, sch Schema) {
	isch, ok := sch.ToInteger()
	if !ok {
		return
	}
	samples := sch.Samples()
	if samples <= 5 {
		return
	}
	seenMin := *isch.SeenMinimum()
	seenMax := *isch.SeenMaximum()
	if seenMin == 0 && seenMax == 1 {
		sch[P_ENUM] = []int{0, 1}
		sch[P_FORMAT] = F_ZERO_ONE
		return
	}
	delete(sch, P_ENUM)
	return
}

// MergeFormat coerces two formats in the 'lowest common denominator'
// (for example, a float and double into a float).
//
// Formats should always be 'compatible' types
// (that is, SniffType is the same for each of the two original values).
//
// Return F_NOFORMAT if the types cannot be coerced, like a UUID and an email.
func MergeFormat(f1, f2 JsonFormat) JsonFormat {
	if f1 == f2 {
		return f1
	}
	if level1, ok := formatCoercions[f1]; ok {
		if level2, ok := level1[f2]; ok {
			return level2
		}
	}
	if level1, ok := formatCoercions[f2]; ok {
		if level2, ok := level1[f1]; ok {
			return level2
		}
	}
	return F_NOFORMAT
}

func mergeSliceProperty(ctx context.Context, f Field, s1, s2 Schema) ([]Schema, bool) {
	flat := make([]Schema, 0, 2)
	if s, ok := s1[f]; ok {
		flat = append(flat, CoerceSlice(s)...)
	} else {
		flat = append(flat, s1)
	}
	if s, ok := s2[f]; ok {
		flat = append(flat, CoerceSlice(s)...)
	} else {
		flat = append(flat, s2)
	}
	accum := make(map[string]Schema, len(flat))
	for _, s := range flat {
		key := fmt.Sprintf("%s-%s", s.Type(), s.Format())
		if entry, ok := accum[key]; ok {
			accum[key] = Merge(ctx, MergeInput{S1: entry, S2: s}).Schema
		} else {
			accum[key] = s
		}
	}
	uniq := make([]Schema, 0, len(accum))
	for _, sch := range accum {
		uniq = append(uniq, sch)
	}
	return uniq, len(uniq) == len(flat)
}

var formatCoercions = (func() map[JsonFormat]map[JsonFormat]JsonFormat {
	m := make(map[JsonFormat]map[JsonFormat]JsonFormat, 20)
	c := func(left JsonFormat, right JsonFormat, to JsonFormat) {
		conv, ok := m[left]
		if !ok {
			conv = make(map[JsonFormat]JsonFormat, 20)
			m[left] = conv
		}
		conv[right] = to
	}
	c(F_DOUBLE, F_DOUBLE, F_DOUBLE)
	c(F_DOUBLE, F_FLOAT, F_DOUBLE)
	c(F_DOUBLE, F_INT32, F_DOUBLE)
	c(F_DOUBLE, F_INT64, F_DOUBLE)
	c(F_DOUBLE, F_TIMESTAMP, F_DOUBLE)
	c(F_DOUBLE, F_TIMESTAMP_MS, F_DOUBLE)
	c(F_DOUBLE, F_ZERO_ONE, F_DOUBLE)

	c(F_FLOAT, F_DOUBLE, F_DOUBLE)
	c(F_FLOAT, F_FLOAT, F_FLOAT)
	c(F_FLOAT, F_INT32, F_FLOAT)
	c(F_FLOAT, F_INT64, F_DOUBLE)
	c(F_FLOAT, F_TIMESTAMP, F_FLOAT)
	c(F_FLOAT, F_TIMESTAMP_MS, F_FLOAT)
	c(F_FLOAT, F_ZERO_ONE, F_FLOAT)

	c(F_INT32, F_DOUBLE, F_DOUBLE)
	c(F_INT32, F_FLOAT, F_FLOAT)
	c(F_INT32, F_INT32, F_INT32)
	c(F_INT32, F_INT64, F_INT64)
	c(F_INT32, F_TIMESTAMP, F_INT32)
	c(F_INT32, F_TIMESTAMP_MS, F_INT32)
	c(F_INT32, F_ZERO_ONE, F_INT32)

	c(F_INT64, F_DOUBLE, F_DOUBLE)
	c(F_INT64, F_FLOAT, F_DOUBLE)
	c(F_INT64, F_INT32, F_INT64)
	c(F_INT64, F_INT64, F_INT64)
	c(F_INT64, F_TIMESTAMP, F_INT64)
	c(F_INT64, F_TIMESTAMP_MS, F_INT64)
	c(F_INT64, F_ZERO_ONE, F_INT64)

	c(F_TIMESTAMP, F_DOUBLE, F_DOUBLE)
	c(F_TIMESTAMP, F_FLOAT, F_FLOAT)
	c(F_TIMESTAMP, F_INT32, F_INT32)
	c(F_TIMESTAMP, F_INT64, F_INT64)
	c(F_TIMESTAMP, F_TIMESTAMP, F_TIMESTAMP)
	c(F_TIMESTAMP, F_TIMESTAMP_MS, F_DOUBLE)
	c(F_TIMESTAMP, F_ZERO_ONE, F_DOUBLE)

	c(F_TIMESTAMP_MS, F_DOUBLE, F_DOUBLE)
	c(F_TIMESTAMP_MS, F_FLOAT, F_FLOAT)
	c(F_TIMESTAMP_MS, F_INT32, F_INT32)
	c(F_TIMESTAMP_MS, F_INT64, F_INT64)
	c(F_TIMESTAMP_MS, F_TIMESTAMP, F_DOUBLE)
	c(F_TIMESTAMP_MS, F_TIMESTAMP_MS, F_TIMESTAMP_MS)
	c(F_TIMESTAMP_MS, F_ZERO_ONE, F_DOUBLE)

	return m
})()

var timeFormatValueComparers = map[JsonFormat]func(values ...string) (nmin, nmax string){
	F_DATE: func(values ...string) (string, string) {
		tmin, tmax := timestring.MinMax(timestring.Many(TF_DATE, values...))
		return tmin.S, tmax.S
	},
	F_DATETIME: func(values ...string) (string, string) {
		tmin, tmax := timestring.MinMax(timestring.Many(TF_DATETIME, values...))
		return tmin.S, tmax.S
	},
	F_DATETIME_NOTZ: func(values ...string) (string, string) {
		tmin, tmax := timestring.MinMax(timestring.Many(TF_DATETIME_NOTZ, values...))
		return tmin.S, tmax.S
	},
	F_TIME: func(values ...string) (string, string) {
		tmin, tmax := timestring.MinMax(timestring.Many(TF_TIME, values...))
		return tmin.S, tmax.S
	},
	F_DURATION: func(values ...string) (string, string) {
		tstrings := make([]timestring.TimeString, len(values))
		for i, s := range values {
			tstrings[i] = timestring.FromPeriod(s)
		}
		tmin, tmax := timestring.MinMax(tstrings)
		return tmin.S, tmax.S
	},
}

type MergeManyInput struct {
	// The 'starting schema'.
	Schema Schema
	// These payloads are merged into the starting schema.
	PayloadIterator moxio.Iterator
	// How many examples to record. Examples are only recorded when there is a significant schema change,
	// like a type changes (new object properties are not recorded).
	// If nil, do not modify examples. If <= 0, delete examples. If > 0, keep only that many examples
	// (total examples are randomly sampled to achieve ExampleLimit examples).
	ExampleLimit *int
}

type MergeManyOutput struct {
	Schema Schema
}

// MergeMany merges payloads into a Schema.
// It can optionally record examples.
// If you only need to process one payload, use MergeOne.
func MergeMany(ctx context.Context, in MergeManyInput) (MergeManyOutput, error) {
	result := in.Schema
	var newExamples []interface{}
	for in.PayloadIterator.Next() {
		msg, err := in.PayloadIterator.Read(ctx)
		if err != nil {
			return MergeManyOutput{Schema: result}, errors.Wrap(err, "payload loader iterator")
		}
		newSchema := Derive("", msg)
		mout := Merge(ctx, MergeInput{Key: "", S1: result, S2: newSchema})
		result = mout.Schema
		if mout.TypeChanged {
			newExamples = append(newExamples, msg)
		}
	}
	if in.ExampleLimit == nil {
	} else if *in.ExampleLimit > 0 {
		result[P_EXAMPLES] = fp.SampleOut(append(Examples(result), newExamples...), *in.ExampleLimit)
	} else {
		delete(result, P_EXAMPLES)
	}
	return MergeManyOutput{Schema: result}, nil
}

type MergeOneInput struct {
	Schema       Schema
	Payload      interface{}
	ExampleLimit *int
}

type MergeOneOutput MergeManyOutput

func MergeOne(ctx context.Context, in MergeOneInput) (MergeOneOutput, error) {
	out, err := MergeMany(ctx, MergeManyInput{
		Schema:          in.Schema,
		ExampleLimit:    in.ExampleLimit,
		PayloadIterator: moxio.NewMemoryIterator([]interface{}{in.Payload}),
	})
	return MergeOneOutput(out), err

}
