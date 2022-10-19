package cmd

import (
	"github.com/lithictech/moxpopuli/moxio"
	"github.com/lithictech/moxpopuli/schemamerge"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var schemagenCmd = &cli.Command{
	Name:  "schemagen",
	Usage: "Given an existing schema, and a series of payloads, fit the schema to the payloads and re-save it.",
	Flags: append(
		append(loaderArgs, saverArgs...),
		&cli.StringFlag{
			Name:     "payload-loader",
			Aliases:  s1("p"),
			Required: true,
			Usage: "Name of the payload loader routine, like 'postgres://x:y@localhost:5432/mydb'. " +
				"Use '-' to read from stdin, or a space to parse payload-loader-arg as the payload. " +
				"See README -> Iterator Loaders for more info.",
		},
		&cli.StringFlag{
			Name:    "payload-loader-arg",
			Aliases: s1("pa"),
			Usage: "Value to pass to the payload loader routine, like a SQL query. " +
				"See README -> Iterator Loaders for more info.",
		},
		examplesFlag,
	),
	Action: func(c *cli.Context) error {
		ctx, _ := newCtx()
		sch, err := loadSchema(ctx, c)
		if err != nil {
			return err
		}

		payloadIterator, err := moxio.LoadIterator(ctx, c.String("payload-loader"), c.String("payload-loader-arg"))
		if err != nil {
			return errors.Wrap(err, "payload loader iterator")
		}

		schout, err := schemamerge.MergeMany(ctx, schemamerge.MergeManyInput{
			Schema:          sch,
			PayloadIterator: payloadIterator,
			ExampleLimit:    examplesValue(c),
		})
		if err != nil {
			return errors.Wrap(err, "merging schemas")
		}

		return save(ctx, c, schout.Schema)
	},
}
