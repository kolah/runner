package rpc

import (
	"fmt"
	"github.com/kolah/runner/internal/app"
	"github.com/kolah/runner/internal/pkg/simplerpc"
	"net"
	"os"
)

func StopHandler() simplerpc.ServerHandlerFunc {
	return func(c net.Conn, args []string) {
		_, err := fmt.Fprintln(c, ServerOK)
		if err != nil {
			fmt.Println("Error returning response")
			//noinspection ALL
			c.Close()
			return
		}
		fmt.Println("Received STOP command")
		os.Exit(0)
	}
}

func SetModeHandler(runner *app.Runner) simplerpc.ServerHandlerFunc {
	return func(c net.Conn, args []string) {
		if len(args) != 1 {
			_, err := fmt.Fprintln(c, ServerErr, "Invalid number of arguments")
			if err != nil {
				fmt.Println("Error returning response")
				return
			}
			fmt.Println("Invalid number of arguments")

			return
		}

		mode := app.RunnerMode(args[0])
		if mode == app.RunnerModeDebug || mode == app.RunnerModeLiveRebuild {
			if err := runner.Build(); err != nil {
				//noinspection ALL
				fmt.Fprintln(c, ServerErr, "Build error")
				//noinspection ALL
				return
			}
			runner.SetMode(mode)
			//noinspection ALL
			fmt.Fprintln(c, ServerOK, "Switched mode to", mode)
			return
		} else {
			//noinspection ALL
			fmt.Fprintln(c, ServerErr, "Unknown mode", mode)
		}

	}
}
