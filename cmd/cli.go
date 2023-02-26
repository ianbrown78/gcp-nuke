package cmd

import (
	"log"
	"os"

	"github.com/ianbrown78/gcp-nuke/config"
	"github.com/ianbrown78/gcp-nuke/resources"
	"github.com/urfave/cli/v2"
)

// Command -
func Command() {

	app := &cli.App{
		Usage:     "The GCP project cleanup tool with added radiation",
		Version:   "v0.1.0",
		UsageText: "e.g. resources-nuke --project resources-nuke-test --dryrun",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "project, p",
				Usage:    "GCP project id to nuke (required)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "no-dryrun, d",
				Usage:    "Do not perform a dryrun",
				Required: false,
			},
			&cli.IntFlag{
				Name:     "timeout, t",
				Value:    400,
				Usage:    "Timeout for removal of a single resource in seconds",
				Required: false,
			},
			&cli.IntFlag{
				Name:     "polltime, pt",
				Value:    10,
				Usage:    "Time for polling resource deletion status in seconds",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "no-keep-project, k",
				Usage:    "Do not keep the project. Delete it with its resources.",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			// Behaviour to delete all resource in parallel in one project at a time - will be made into loop / concurrenct project nuke if required
			config := config.Config{
				Project:       c.String("project"),
				NoDryRun:      c.Bool("no-dryrun"),
				Timeout:       c.Int("timeout"),
				PollTime:      c.Int("polltime"),
				NoKeepProject: c.Bool("no-keep-project"),
				Context:       resources.Ctx,
				Zones:         resources.GetZones(resources.Ctx, c.String("project")),
				Regions:       resources.GetRegions(resources.Ctx, c.String("project")),
			}

			log.Printf("[Info] Timeout %v seconds. Polltime %v seconds. Dry run: %v", config.Timeout, config.PollTime, config.NoDryRun)
			resources.RemoveProjectResources(config)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
