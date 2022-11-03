package fixturegen

import (
	"encoding/json"
	"github.com/lithictech/moxpopuli/faker"
	. "github.com/lithictech/moxpopuli/jsonformat"
	"github.com/lithictech/moxpopuli/schema"
	"math"
	"strconv"
	"time"
)

type RunInput struct {
	Count int
	TZ    *time.Location
}

func Run(in RunInput) []map[string]interface{} {
	result := make([]map[string]interface{}, in.Count)
	for i := 0; i < in.Count; i++ {
		result[i] = Generate(GenerateInput{TZ: in.TZ})
	}
	return result
}

type GenerateInput struct {
	TZ *time.Location
}

func Generate(in GenerateInput) map[string]interface{} {
	if in.TZ == nil {
		in.TZ = time.Local
	}
	m := map[JsonFormat]interface{}{
		F_DOUBLE:        faker.Float64(math.MaxFloat32+1, math.MaxFloat32+1_000_000_000_000),
		F_FLOAT:         faker.Float64(-1*(math.MaxFloat32-1), math.MaxFloat32-1),
		F_INT32:         faker.Int64(math.MinInt32+1, math.MaxInt32-1),
		F_INT64:         faker.Int64(math.MaxInt32+1, math.MaxInt64),
		F_TIMESTAMP:     faker.Int64(0, time.Now().In(in.TZ).AddDate(10, 0, 0).Unix()),
		F_TIMESTAMP_MS:  faker.Int64(0, time.Now().In(in.TZ).AddDate(10, 0, 0).UnixMilli()),
		F_ZERO_ONE:      faker.Choice(0, 1),
		F_BINARY:        faker.Bytes(),
		F_BYTE:          faker.Base64(faker.Hex()),
		F_EMAIL:         faker.Email(),
		F_COUNTRY:       faker.Country(),
		F_CURRENCY:      faker.Currency(),
		F_IPV4:          faker.IPv4(),
		F_IPV6:          faker.IPv6(),
		F_URI:           faker.URL().String(),
		F_UUID4:         faker.UUID4(),
		F_NUMERICAL:     strconv.Itoa(faker.Int()),
		F_DATE:          faker.Time().Format(schema.TF_DATE),
		F_DATETIME:      faker.Time().Format(schema.TF_DATETIME),
		F_DATETIME_NOTZ: faker.Time().Format(schema.TF_DATETIME_NOTZ),
		F_TIME:          faker.Time().Format(schema.TF_TIME),
		F_DURATION:      faker.Period().String(),
		"identifier":    faker.Hex(12),
		"noformat":      faker.Choice(faker.Int64(), faker.Bool(), faker.Float64(), faker.Hex()),
		"object_uuid": map[string]interface{}{
			string(F_UUID4): faker.UUID4(),
		},
		"array_uuid": []interface{}{
			faker.UUID4(),
			faker.UUID4(),
		},
		"array_object": []interface{}{
			map[string]interface{}{
				string(F_UUID4): faker.UUID4(),
			},
		},
		"object_array": map[string]interface{}{
			"array_uuid": []interface{}{
				faker.UUID4(),
				faker.UUID4(),
			},
		},
	}

	// We need to round trip to get standard JSON types as if it were really loaded.
	b, err := json.Marshal(m)
	if err != nil {
		panic("should never have errored marshaling: " + err.Error())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		panic("should never have errored unmarshaling: " + err.Error())
	}
	return result
}
