package jsonformat

import (
	"encoding/base64"
	"github.com/lithictech/moxpopuli/jsontype"
	"github.com/rickb777/date/period"
	"math"
	"net"
	"regexp"
	"strings"
	"time"
)

type JsonFormat string

//goland:noinspection GoSnakeCaseUsage
const (
	F_DOUBLE       JsonFormat = "double"
	F_FLOAT        JsonFormat = "float"
	F_INT32        JsonFormat = "int32"
	F_INT64        JsonFormat = "int64"
	F_TIMESTAMP    JsonFormat = "timestamp"
	F_TIMESTAMP_MS JsonFormat = "timestamp-ms"
	F_ZERO_ONE     JsonFormat = "zero-one"

	F_BINARY    JsonFormat = "binary"
	F_BYTE      JsonFormat = "byte"
	F_EMAIL     JsonFormat = "email"
	F_COUNTRY   JsonFormat = "iso-country"
	F_CURRENCY  JsonFormat = "iso-currency"
	F_IPV4      JsonFormat = "ipv4"
	F_IPV6      JsonFormat = "ipv6"
	F_URI       JsonFormat = "uri"
	F_UUID4     JsonFormat = "uuid4"
	F_NUMERICAL JsonFormat = "numerical"

	F_DATE          JsonFormat = "date"
	F_DATETIME      JsonFormat = "date-time"
	F_DATETIME_NOTZ JsonFormat = "date-time-notz"
	F_TIME          JsonFormat = "time"
	F_DURATION      JsonFormat = "duration"

	F_NOFORMAT JsonFormat = ""
)

func IsChronolike(f JsonFormat) bool {
	return f == F_DATETIME ||
		f == F_DATETIME_NOTZ ||
		f == F_TIME ||
		f == F_DATE ||
		f == F_DURATION
}

// Sniff returns the JsonFormat for the given value.
// Value can be a Go primitive or a special type like json.Number.
// See https://www.asyncapi.com/docs/reference/specification/v2.4.0#dataTypeFormat
// for possible formats.
func Sniff(t jsontype.JsonType, value interface{}) JsonFormat {
	if t == jsontype.T_STRING {
		s := value.(string)
		if sniffEmail(s) {
			return F_EMAIL
		} else if sniffUrl(s) {
			return F_URI
		} else if sniffIPv4(s) {
			return F_IPV4
		} else if sniffIPv6(s) {
			return F_IPV6
		} else if sniffCountry(s) {
			return F_COUNTRY
		} else if sniffCurrency(s) {
			return F_CURRENCY
		} else if sniffUuid4(s) {
			return F_UUID4
		} else if sniffNumericalString(s) {
			return F_NUMERICAL
		} else if sniffDateTimeNoTZ(s) {
			return F_DATETIME_NOTZ
		} else if sniffDateTimeTZ(s) {
			// MUST go after 'noTz' search
			return F_DATETIME
		} else if sniffDate(s) {
			return F_DATE
		} else if sniffTime(s) {
			return F_TIME
		} else if sniffDuration(s) {
			return F_DURATION
		} else if sniffBinary(s) {
			return F_BINARY
		} else if sniffBase64(s) {
			return F_BYTE
		}
		return F_NOFORMAT
	} else if t == jsontype.T_INTEGER {
		// We can fit any integer into a float, and we don't need the *actual* value,
		// so use floats for sniff functions that also need floats.
		i := value.(int)
		f := float64(i)
		if sniffTimestamp(f) {
			return F_TIMESTAMP
		} else if sniffTimestampMS(f) {
			return F_TIMESTAMP_MS
		} else if sniffInt32(i) {
			return F_INT32
		}
		return F_INT64
	} else if t == jsontype.T_NUMBER {
		f := value.(float64)
		if sniffTimestamp(f) {
			return F_TIMESTAMP
		} else if sniffTimestampMS(f) {
			return F_TIMESTAMP_MS
		} else if sniffFloat32(f) {
			return F_FLOAT
		}
		return F_DOUBLE
	}
	return F_NOFORMAT
}

func sniffBinary(_ string) bool {
	// Not sure when we see this.
	return false
}
func sniffBase64(s string) bool {
	if len(s) < 40 {
		// 40 encoding chars is 30 ascii chars, anything less than that is probably not worth encoding.
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func sniffUrl(s string) bool {
	// Do not use URL parsing, it hits tons of false-positives.
	return strings.HasPrefix(s, "/") || urlRegex.MatchString(s)
}

var urlRegex = regexp.MustCompile("^[a-z]+://")

func sniffIPv4(s string) bool {
	p := net.ParseIP(s)
	return p != nil && p.To4() != nil
}

func sniffIPv6(s string) bool {
	p := net.ParseIP(s)
	return p != nil && p.To16() != nil
}

func sniffEmail(s string) bool {
	return emailRegex.MatchString(s)
}

var emailRegex = regexp.MustCompile("^.*@([a-zA-Z0-9-]+\\.)+[a-zA-Z]{2,6}$")

func sniffCountry(s string) bool {
	return countryRegex.MatchString(s)
}

var countryRegex = regexp.MustCompile("^[A-Z][A-Z]$")

func sniffCurrency(s string) bool {
	return currencyRegex.MatchString(s)
}

var currencyRegex = regexp.MustCompile("^([A-Z]{3}|[a-z]{3})$")

func sniffUuid4(s string) bool {
	return uuid4Regex.MatchString(s)
}

var uuid4Regex = regexp.MustCompile("^[0-9a-fA-F]{8}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{12}$")

func sniffNumericalString(s string) bool {
	return numericalStrRegex.MatchString(s)
}

var numericalStrRegex = regexp.MustCompile("^-?\\d+$")

func sniffDateTimeTZ(s string) bool {
	return dttzRegex.MatchString(s)
}

func sniffDateTimeNoTZ(s string) bool {
	return dtNoTzRegex.MatchString(s)
}

var dttzRegex = regexp.MustCompile("^[0-9][0-9][0-9][0-9]-[0-1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9].?\\d*(Z|[+-]\\d{2}:?\\d{2})?$")
var dtNoTzRegex = regexp.MustCompile("^[0-9][0-9][0-9][0-9]-[0-1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9]\\.?\\d*$")

func sniffDate(s string) bool {
	return dateRegex.MatchString(s)
}

var dateRegex = regexp.MustCompile("^[0-9][0-9][0-9][0-9]-[0-1][0-9]-[0-3][0-9]$")

func sniffTime(s string) bool {
	return timeRegex.MatchString(s)
}

var timeRegex = regexp.MustCompile("^[0-2][0-9]:[0-5][0-9]:[0-5][0-9].?\\d*([+-]\\d{2}:?\\d{2})?$")

func sniffDuration(s string) bool {
	_, err := period.Parse(s)
	return err == nil
}

func sniffTimestamp(f float64) bool {
	return sniffTimestampMS(f * 1000)
}

func sniffTimestampMS(f float64) bool {
	upperBound := float64(time.Now().Add(circaYear * 40).UnixMilli())
	lowerBound := float64(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli())
	return f > lowerBound && f < upperBound
}

const circaYear = time.Hour * 24 * 365

func sniffInt32(f int) bool {
	return f >= math.MinInt32 && f <= math.MaxInt32
}

func sniffFloat32(f float64) bool {
	return f >= minFloat32 && f <= maxFloat32
}

// copied from Go 1.19
const maxFloat32 = 0x1p127 * (1 + (1 - 0x1p-23))
const minFloat32 = -1 * maxFloat32
