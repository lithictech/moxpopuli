package cmd

import (
	"github.com/lithictech/moxpopuli/moxvox"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"io"
	"regexp"
)

var voxCmd = &cli.Command{
	Name:  "vox",
	Usage: "Publish events against channels in the given spec.",
	Flags: append(
		loaderArgs,
		&cli.StringFlag{
			Name:  "match",
			Value: ".*",
			Usage: "Regular expression to use against subscriber channel names. Only those that match get exercised.",
		},
		&cli.BoolFlag{
			Name:  "print",
			Usage: "If given, dump requests and responses to stdout.",
		},
		countFlag,
		bindingFlag,
	),
	Action: func(c *cli.Context) error {
		ctx := newCtx()
		apispec, err := loadMap(ctx, c)
		if err != nil {
			return err
		}
		var voxer moxvox.Vox
		if c.String("binding") == "http" {
			voxer = moxvox.HttpVox
		} else {
			return errors.New("unsupported binding")
		}
		matcher, err := regexp.Compile(c.String("match"))
		if err != nil {
			return errors.Wrap(err, "invalid match regex")
		}
		var printer io.Writer
		if c.Bool("print") {
			printer = c.App.Writer
		}
		return voxer(ctx, moxvox.VoxInput{
			Spec:           apispec,
			Count:          c.Int("count"),
			ChannelMatcher: matcher,
			Printer:        printer,
		})
	},
}
