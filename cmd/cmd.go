package cmd

import (
	"context"
	"fmt"
	"github.com/lithictech/go-aperitif/logctx"
	"github.com/lithictech/moxpopuli"
	"github.com/lithictech/moxpopuli/moxio"
	"github.com/lithictech/moxpopuli/schema"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func Execute() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "debug", Value: false},
		},
		Commands: []*cli.Command{
			schemagenCmd,
			datagenCmd,
			fixtureGenCmd,
			specgenCmd,
			voxCmd,
			serverCmd,
			{
				Name: "version",
				Action: func(c *cli.Context) error {
					fmt.Fprintln(os.Stdout, moxpopuli.BuildSha[0:8])
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func newCtx() (context.Context, moxpopuli.Config) {
	cfg := moxpopuli.LoadConfig()
	logger, err := logctx.NewLogger(logctx.NewLoggerInput{
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		File:      cfg.LogFile,
		BuildSha:  moxpopuli.BuildSha,
		BuildTime: moxpopuli.BuildTime,
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	ctx = logctx.WithLogger(ctx, logger)
	return logctx.WithTracingLogger(ctx), cfg
}

var loaderArgs = []cli.Flag{
	&cli.StringFlag{
		Name:    "loader",
		Aliases: s1("l"),
		Usage: "Name of the loader routine, like 'file://./temp/myschema.json'. " +
			"If empty, do not load data. " +
			"See README -> Single Objects Load and Save for more info.",
	},
	&cli.StringFlag{
		Name:    "loader-arg",
		Aliases: s1("la"),
		Usage: "Value to pass to the loader, like a JSON path. " +
			"See README -> Single Objects Load and Save for more info.",
	},
}

func loadSchema(ctx context.Context, c *cli.Context) (schema.Schema, error) {
	return schema.Load(ctx, c.String("loader"), c.String("loader-arg"))
}

func loadMap(ctx context.Context, c *cli.Context) (map[string]interface{}, error) {
	return moxio.LoadOneMap(ctx, c.String("loader"), c.String("loader-arg"))
}

var saverArgs = []cli.Flag{
	&cli.StringFlag{
		Name:    "saver",
		Aliases: s1("s"),
		Usage: "Name of the saver routine, like 'file://./temp/myschema.json'. " +
			"Write to stdout if empty. See README -> Savers for more.",
	},
	&cli.StringFlag{
		Name:    "saver-arg",
		Aliases: s1("sa"),
		Usage:   "Value to pass to the saver, like a JSON path. See README -> Savers for more.",
	},
}

func save(ctx context.Context, c *cli.Context, i interface{}) error {
	if err := moxio.Save(ctx, c.String("saver"), c.String("saver-arg"), i); err != nil {
		return errors.Wrap(err, "saving")
	}
	return nil
}

var bindingFlag = &cli.StringFlag{
	Name:    "binding",
	Aliases: s1("b"),
	Value:   "http",
	Usage:   "Binding for the service. Determines how events are interpreted. See README for more info.",
}

var countFlag = &cli.IntFlag{
	Name:  "count",
	Value: 1,
	Usage: "Number of payloads to generate.",
}

func s1(s string) []string {
	return []string{s}
}

var examplesFlag = &cli.IntFlag{
	Name:    "examples",
	Aliases: s1("e"),
	Usage: "If given, record up to this many examples that modify the schema. " +
		"If not given or <= 0, do not record examples.",
}

func examplesValue(c *cli.Context) *int {
	var examples int
	if c.IsSet("examples") {
		examples = c.Int("examples")
	}
	return &examples
}
