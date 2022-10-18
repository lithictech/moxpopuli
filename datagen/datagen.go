package datagen

import (
	"fmt"
	"github.com/lithictech/moxpopuli/faker"
	. "github.com/lithictech/moxpopuli/jsonformat"
	. "github.com/lithictech/moxpopuli/schema"
	"github.com/lithictech/moxpopuli/timestring"
	"github.com/rickb777/date/period"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Generate(key string, sch Schema) interface{} {
	f := sch.Format()
	// Remember that 'seen min' and 'seen max' will always be valid for the format,
	// so we can use int64 and float64 fakes and be sure we're getting int32, etc.
	if scht, ok := sch.ToInteger(); ok {
		if f == F_ZERO_ONE {
			return faker.Choice([]interface{}{0, 1})
		}
		return faker.Int(*scht.SeenMinimum(), *scht.SeenMaximum())
	} else if scht, ok := sch.ToNumber(); ok {
		return faker.Float64(*scht.SeenMinimum(), *scht.SeenMaximum())
	} else if _, ok := sch.ToBoolean(); ok {
		return faker.Bool()
	} else if scht, ok := sch.ToString(); ok {
		if len(scht.Enum()) > 0 {
			return faker.ChoiceString(scht.Enum()...)
		}
		switch f {
		case F_BINARY:
			return string(faker.Bytes(*scht.SeenMinLength(), *scht.SeenMaxLength()))
		case F_BYTE:
			return faker.Base64(faker.Hex(*scht.SeenMinLength(), *scht.SeenMaxLength()))
		case F_EMAIL:
			return faker.Email()
		case F_COUNTRY:
			return faker.Country()
		case F_CURRENCY:
			return faker.Currency()
		case F_IPV4:
			return faker.IPv4()
		case F_IPV6:
			return faker.IPv6()
		case F_URI:
			loc := faker.ChoiceString(scht.SeenUriLocations()...)
			locUrl, _ := url.Parse(loc)
			pathUrl := faker.URL()
			locUrl.Path = pathUrl.Path
			locUrl.RawQuery = pathUrl.RawQuery
			return locUrl.String()
		case F_UUID4:
			return faker.UUID4()
		case F_NUMERICAL:
			min, _ := strconv.Atoi(*scht.SeenMinimum())
			max, _ := strconv.Atoi(*scht.SeenMaximum())
			fmt.Println(min, max)
			return strconv.Itoa(faker.Int(min, max))
		case F_DATE:
			return timeFaker(key, scht, TF_DATE)
		case F_DATETIME:
			return timeFaker(key, scht, TF_DATETIME)
		case F_DATETIME_NOTZ:
			return timeFaker(key, scht, TF_DATETIME_NOTZ)
		case F_TIME:
			return timeFaker(key, scht, TF_TIME)
		case F_DURATION:
			tmin, tmax := timestring.FromPeriod(*scht.SeenMinimum()), timestring.FromPeriod(*scht.SeenMaximum())
			d := faker.Int64(tmin.U, tmax.U)
			p, _ := period.NewOf(time.Duration(d))
			return p.Format()
		default:
			return faker.Hex(*scht.SeenMinLength(), *scht.SeenMaxLength())
		}
	} else if scht, ok := sch.ToArray(); ok {
		arr := make([]interface{}, faker.Int(*scht.SeenMinLength(), *scht.SeenMaxLength()))
		for i := range arr {
			arr[i] = Generate(strconv.Itoa(i), scht.Items())
		}
		return arr
	} else if scht, ok := sch.ToObject(); ok {
		r := make(map[string]interface{}, len(scht.Properties()))
		for k, v := range scht.Properties() {
			r[k] = Generate(k, v)
		}
		return r
	} else if sch.Nullable() {
		return nil
	} else {
		return faker.Choice([]interface{}{faker.Hex(), faker.Int()})
	}
}

func timeFaker(key string, sch StringSchema, layout string) string {
	tmin, tmax := timestring.From(layout, *sch.SeenMinimum()), timestring.From(layout, *sch.SeenMaximum())
	// 'updated at' should be possible to create going forward,
	// otherwise we won't be able to run updates ever.
	if strings.HasPrefix(key, "updated") || strings.HasPrefix(key, "modified") {
		tmax = timestring.From(layout, time.Now().Format(layout))
	}
	return faker.Time(tmin.T, tmax.T).Format(layout)
}
