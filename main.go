package main

import (
	"go-api/server"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	apiServerAddrFlagName string = "addr"
)

func main() {
	if err := app().Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("could not run application")
	}
}

func app() *cli.App {
	return &cli.App{
		Name:  "go-api",
		Usage: "The API",
		Commands: []*cli.Command{
			apiServerCommand(),
		},
	}
}

func apiServerCommand() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "starts the API server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    apiServerAddrFlagName,
				EnvVars: []string{"API_SERVER_ADDR"},
			},
		},
		Action: func(c *cli.Context) error {
			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			stoppper := make(chan struct{})
			go func() {
				<-done
				close(stoppper)
			}()
			addr := c.String(apiServerAddrFlagName)
			srv, err := server.NewServer(addr)
			if err != nil {
				return err
			}
			return srv.Start(stoppper)
		},
	}
}
