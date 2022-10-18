package redact

import (
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"encoding/json"
	"github.com/AlessandroPomponio/go-gibberish/gibberish"
	"github.com/AlessandroPomponio/go-gibberish/structs"
	"github.com/lithictech/moxpopuli/internal"
	"math/rand"
	"regexp"
	"sync"
)

func Zero(s string) string {
	runes := make([]rune, len(s))
	for i, r := range s {
		if r >= ca && r <= cz {
			runes[i] = 'a'
		} else if r >= cA && r <= cZ {
			runes[i] = 'A'
		} else if r >= c0 && r <= c9 {
			runes[i] = '0'
		} else {
			runes[i] = r
		}
	}
	return string(runes)
}

func Randomize(s string) string {
	runes := make([]rune, len(s))
	for i, r := range s {
		if r >= ca && r <= cz {
			runes[i] = randRune(ca, cz)
		} else if r >= cA && r <= cZ {
			runes[i] = randRune(cA, cZ)
		} else if r >= c0 && r <= c9 {
			runes[i] = randRune(c0, c9)
		} else {
			runes[i] = r
		}
	}
	return string(runes)
}

func randRune(lower, upper rune) rune {
	diff := upper - lower
	c := rand.Int31n(diff + 1)
	return lower + c
}

const ca = 'a'
const cz = 'z'
const cA = 'A'
const cZ = 'Z'
const c0 = '0'
const c9 = '9'

func IsGibberish(s string) bool {
	loadGibberish.Do(func() {
		b, err := fs.ReadFile("packaged-knowledge.json")
		internal.Assert(err == nil, "should never fail to load embedded file")
		err = json.Unmarshal(b, &gibberishData)
		internal.Assert(err == nil, "should never fail to load embedded file")
	})
	return gibberish.IsGibberish(s, &gibberishData)
}

var loadGibberish = sync.Once{}

var gibberishData structs.GibberishData

//go:embed packaged-knowledge.json
var fs embed.FS

func IsSha(x string) bool {
	return shaRegex.MatchString(x)
}

var shaRegex = regexp.MustCompile("^[a-z0-9]{40}$")

func UnsafeVariableHash(data, salt []byte) string {
	h := sha512.New()
	h.Write(data)
	h.Write(salt)
	sum := h.Sum(nil)
	s := unsafeBase64.EncodeToString(sum)
	if len(data) == len(s) {
		return s
	}
	if len(data) < len(s) {
		return s[:len(data)]
	}
	s2 := s + s[:len(s)-len(data)]
	return s2
}

var unsafeBase64 = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789ab").WithPadding(base64.NoPadding)
