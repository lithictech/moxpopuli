package cmd

import (
	"github.com/lithictech/moxpopuli/asyncapispecmerge"
	"github.com/lithictech/moxpopuli/moxio"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var specgenCmd = &cli.Command{
	Name:  "specgen",
	Usage: "Generate an entire AsyncAPI specification based on events.",
	Flags: append(
		append(loaderArgs, saverArgs...),
		&cli.StringFlag{
			Name:     "event-loader",
			Aliases:  s1("e"),
			Required: true,
			Usage: "Name of the event loader routine, like 'postgres://x:y@localhost:5432/mydb'. " +
				"Use '-' to read from stdin, or '_' to parse event-loader-arg as the event. " +
				"See README -> Iterator Loaders for more info.",
		},
		&cli.StringFlag{
			Name:    "event-loader-arg",
			Aliases: s1("ea"),
			Usage: "Value to pass to the event loader routine, like a SQL query. " +
				"See README -> Iterator Loaders for more info.",
		},
		bindingFlag,
	),
	Action: func(c *cli.Context) error {
		ctx := newCtx()
		spec, err := loadMap(ctx, c)
		if err != nil {
			return err
		}
		iter, err := moxio.LoadIterator(ctx, c.String("event-loader"), c.String("event-loader-arg"))
		if err != nil {
			return errors.Wrap(err, "loader iterator")
		}
		var merge asyncapispecmerge.Merge
		if c.String("binding") == "http" {
			merge = asyncapispecmerge.MergeHttp
		} else {
			return errors.New("unsupported binding")
		}
		if err := merge(ctx, asyncapispecmerge.MergeInput{
			Spec:          spec,
			EventIterator: iter,
		}); err != nil {
			return errors.Wrap(err, "merging")
		}

		return save(ctx, c, spec)
	},
}
