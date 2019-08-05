package cli

import (
	"fmt"
	"github.com/kolah/runner/internal/app"
	"github.com/kolah/runner/internal/app/config"
	"github.com/kolah/runner/internal/app/rpc"
	"github.com/kolah/runner/internal/pkg/simplerpc"
	"log"

	"github.com/spf13/cobra"
	"os"
)

var controlCmd = &cobra.Command{
	Use:   "ctl [debug|rebuild|stop]",
	Short: "Allows to set runner mode",

	Run: func(cmd *cobra.Command, args []string) {
		configuration, err := config.LoadConfig(cmd)
		if err != nil {
			log.Fatal("Failed to load config: " + err.Error())
		}
		if len(args) != 1 {
			_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Invalid number of arguments")
			os.Exit(1)
		}

		c := simplerpc.NewClient("localhost", configuration.CtlPort)

		if err := c.Connect(); err != nil {
			_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Dial error", err)
			os.Exit(1)
		}

		//noinspection ALL
		defer c.Close()

		var msg string

		mode := args[0]
		switch mode {
		case "debug":
			//noinspection ALL
			fmt.Fprintln(cmd.OutOrStdout(), "Switching runner to debug mode")
			msg = fmt.Sprintf("%s %s", rpc.ClientSetMode, app.RunnerModeDebug)
			break
		case "rebuild":
			//noinspection ALL
			fmt.Fprintln(cmd.OutOrStdout(), "Switching runner to live rebuild mode")
			msg = fmt.Sprintf("%s %s", rpc.ClientSetMode, app.RunnerModeLiveRebuild)
			break
		case "stop":
			//noinspection ALL
			fmt.Fprintln(cmd.OutOrStdout(), "Stopping runner")
			msg = fmt.Sprintf("%s", rpc.ClientStop)
			break
		default:
			//noinspection ALL
			fmt.Fprintln(cmd.OutOrStderr(), "Invalid mode", mode)
			os.Exit(1)
		}

		//noinspection ALL
		fmt.Fprintln(cmd.OutOrStdout(), "Client sent:", msg)
		line, err := c.SendCommand(msg)
		if err != nil {
			//noinspection ALL
			fmt.Fprintln(cmd.OutOrStderr(), "Communication error:", err)
			os.Exit(1)
		}

		//noinspection ALL
		fmt.Fprintln(cmd.OutOrStdout(), "Response:", line)
	},

}


