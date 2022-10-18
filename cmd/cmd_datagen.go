package cmd

import (
	"github.com/lithictech/moxpopuli/datagen"
	"github.com/lithictech/moxpopuli/moxjson"
	"github.com/urfave/cli/v2"
)

var datagenCmd = &cli.Command{
	Name:  "datagen",
	Usage: "Generate example payloads from the loaded schema.",
	Flags: append(
		loaderArgs,
		&cli.IntFlag{
			Name:  "count",
			Value: 1,
			Usage: "Number of payloads to generate.",
		}),
	Action: func(c *cli.Context) error {
		ctx := newCtx()
		schema, err := loadSchema(ctx, c)
		if err != nil {
			return err
		}
		enc := moxjson.NewPrettyEncoder(c.App.Writer)
		for i := 0; i < c.Int("count"); i++ {
			pl := datagen.Generate("", schema)
			if err := enc.Encode(pl); err != nil {
				return err
			}
		}
		return nil
	},
}
