package cmd

import (
	"encoding/json"
	"github.com/lithictech/moxpopuli/fixturegen"
	"github.com/lithictech/moxpopuli/moxjson"
	"github.com/urfave/cli/v2"
)

var fixtureGenCmd = &cli.Command{
	Name:  "fixturegen",
	Usage: "Write JSON to stdout, that contains fields with faked values for all supported formats.",
	Flags: []cli.Flag{
		countFlag,
		&cli.BoolFlag{
			Name:  "lines",
			Usage: "If true, write JSON lines instead of an array object. You can generally pipe --lines output as 'moxpopuli schemagen -p=-'.",
		},
	},
	Action: func(c *cli.Context) error {
		m := fixturegen.Run(fixturegen.RunInput{Count: c.Int("count")})
		if c.Bool("lines") {
			enc := json.NewEncoder(c.App.Writer)
			for _, o := range m {
				if err := enc.Encode(o); err != nil {
					return err
				}
			}
			return nil
		}
		return moxjson.NewPrettyEncoder(c.App.Writer).Encode(m)
	},
}
