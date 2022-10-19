package cmd

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/lithictech/go-aperitif/api"
	"github.com/lithictech/go-aperitif/logctx"
	"github.com/lithictech/moxpopuli"
	"github.com/lithictech/moxpopuli/v1"
	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"strings"
)

var serverCmd = &cli.Command{
	Name:        "server",
	Description: "Run the server",
	Flags: []cli.Flag{
		portFlag,
	},
	Action: func(c *cli.Context) error {
		ctx, cfg := newCtx()
		logger := logctx.Logger(ctx)
		e := echo.New()
		v1.MountSwaggerui(e)
		api.New(api.Config{
			App:    e,
			Logger: logger,
			LoggingMiddlwareConfig: api.LoggingMiddlwareConfig{
				DoLog: func(c echo.Context, e *logrus.Entry) {
					if strings.HasPrefix(c.Request().URL.Path, "/swaggerui") && c.Response().Status < 400 {
						e.Debug("request_finished")
					} else {
						api.LoggingMiddlewareDefaultDoLog(c, e)
					}
				},
			},
			CorsOrigins: cfg.CorsOrigins,
			StatusResponse: map[string]interface{}{
				"build_sha":  moxpopuli.BuildSha,
				"build_time": moxpopuli.BuildTime,
				"message":    "quack",
			},
			HealthHandler: func(c echo.Context) error {
				return c.JSON(200, map[string]interface{}{"o": "k"})
			},
		})
		v1.Register(e)
		port := getPort(c)
		logger.WithField("port", port).Info("server_listening")
		if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
		return nil
	},
	Subcommands: []*cli.Command{
		{
			Name: "openapi",
			Action: func(c *cli.Context) error {
				sash := v1.NewSashay()
				_, err := c.App.Writer.Write([]byte(sash.BuildYAML()))
				return err
			},
		},
		{
			Name:  "swaggerui",
			Flags: []cli.Flag{portFlag},
			Action: func(c *cli.Context) error {
				return browser.OpenURL(fmt.Sprintf("http://localhost:%d/swaggerui/index.html", getPort(c)))
			},
		},
	},
}

var portFlag = &cli.IntFlag{Name: "port", Aliases: s1("p"), EnvVars: []string{"PORT"}, Value: 18022}

func getPort(c *cli.Context) int {
	return c.Int("port")
}
