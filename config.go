package moxpopuli

import (
	"fmt"
	"github.com/lithictech/go-aperitif/convext"
	"os"
	"strings"
	"time"
)

var BuildTime = time.Unix(0, 0).Format(time.RFC3339)
var BuildSha = "00000000"

type Config struct {
	CorsOrigins []string
	LogFile     string
	LogFormat   string
	LogLevel    string
}

func LoadConfig() Config {
	cfg := Config{
		CorsOrigins: strings.Split(mustenvstr("CORS_ORIGINS"), ","),
		LogFile:     os.Getenv("LOG_FILE"),
		LogFormat:   os.Getenv("LOG_FORMAT"),
		LogLevel:    mustenvstr("LOG_LEVEL"),
	}
	return cfg
}

func mustenvstr(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("'%s' should have had a default set, something weird happened", k))
	}
	return v
}

func init() {
	setenv := func(k string, v interface{}) {
		if _, ok := os.LookupEnv(k); !ok {
			convext.Must(os.Setenv(k, fmt.Sprintf("%v", v)))
		}
	}
	setenv("CORS_ORIGINS", "http://localhost:18022,https://*.webhookdb.com,https://webhookdb.com")
	setenv("LOG_LEVEL", "info")
}
