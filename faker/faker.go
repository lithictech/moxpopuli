package faker

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/rickb777/date/period"
	"math/rand"
	"net/url"
	"time"
)

func Hex(strlen ...int) string {
	var slen int
	if len(strlen) == 0 {
		slen = Int(4, 20)
	} else if len(strlen) == 1 {
		slen = strlen[0]
	} else {
		slen = Int(strlen[0], strlen[1])
	}
	bytes := make([]byte, slen/2)
	if _, err := crand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func UUID4() string {
	// ec5e4555-60b0-46f6-90e3-935b34e8c2ba
	return fmt.Sprintf("%s-%s-%s-%s-%s", Hex(8), Hex(4), Hex(4), Hex(4), Hex(12))
}

func Int(minmax ...int) int {
	mm64 := make([]int64, len(minmax))
	for i, m := range minmax {
		mm64[i] = int64(m)
	}
	return int(Int64(mm64...))
}

func Int64(minmax ...int64) int64 {
	min := int64(-1_000_000_000_000)
	max := int64(1_000_000_000_000)
	if len(minmax) >= 2 {
		max = minmax[1]
	}
	if len(minmax) >= 1 {
		min = minmax[0]
	}
	diff := max - min
	if diff == 0 {
		return max
	}
	return min + rand.Int63n(diff)
}

func IPv4() string {
	return faker.IPv4()
}

func IPv6() string {
	return faker.IPv6()
}

func URL() *url.URL {
	u, _ := url.Parse(faker.URL())
	return u
}

func Sign() int {
	if Bool() {
		return 1
	}
	return -1
}

func Float64(minmax ...float64) float64 {
	min := -1_000_000_000_000.0
	max := 1_000_000_000_000.0
	if len(minmax) >= 2 {
		max = minmax[1]
	}
	if len(minmax) >= 1 {
		min = minmax[0]
	}
	diff := max - min
	return min + (rand.Float64() * diff)
}

func Time(minmax ...time.Time) time.Time {
	min := time.Now().AddDate(-10, 0, 0)
	max := time.Now().AddDate(10, 0, 0)
	if len(minmax) >= 2 {
		max = minmax[1]
	}
	if len(minmax) >= 1 {
		min = minmax[0]
	}
	u := Int64(min.UnixMicro(), max.UnixMicro())
	return time.UnixMicro(u)
}

func Period() period.Period {
	p := period.New(Int(0, 5), Int(0, 11), Int(0, 28), Int(0, 23), Int(0, 59), Int(0, 59))
	p = p.Simplify(false)
	return p
}

func Choice(items ...interface{}) interface{} {
	return items[rand.Intn(len(items))]
}

func ChoiceString(items ...string) string {
	return items[rand.Intn(len(items))]
}

func Base64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func Bytes(minmax ...int) []byte {
	if len(minmax) == 0 {
		minmax = []int{4, 12}
	} else if len(minmax) == 1 {
		minmax = []int{minmax[0], 12}
	}
	size := Int64(int64(minmax[0]), int64(minmax[1]))
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func Email() string {
	return faker.Email()
}

func Currency() string {
	return faker.Currency()
}

func Country() string {
	return faker.Currency()[:2]
}

func Bool() bool {
	return rand.Intn(2) == 0 // always 0 or 1
}
