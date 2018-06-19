package cmd

import (
	"github.com/spf13/cobra"
	"net"
	"fmt"
	"bufio"
	"os"
	"github.com/kolah/runner/worker"
	"github.com/kolah/runner/config"
)

var controlCmd = &cobra.Command{
	Use:   "ctl [debug|rebuild|stop]",
	Short: "Allows to set runner mode",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(cmd.OutOrStderr(), "Invalid number of arguments")
			os.Exit(1)
		}

		c, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.Config.RunnerPort))
		if err != nil {
			fmt.Fprintln(cmd.OutOrStderr(), "Dial error", err)
			os.Exit(1)
		}
		defer c.Close()

		var msg string

		mode := args[0]
		switch mode {
		case "debug":
			fmt.Fprintln(cmd.OutOrStdout(), "Switching runner to debug mode")
			msg = fmt.Sprintf("%s %s", ClientSetMode, worker.RunnerModeDebug)
			break
		case "rebuild":
			fmt.Fprintln(cmd.OutOrStdout(), "Switching runner to live rebuild mode")
			msg = fmt.Sprintf("%s %s", ClientSetMode, worker.RunnerModeLiveRebuild)
		case "stop":
			fmt.Fprintln(cmd.OutOrStdout(), "Stopping runner")
			msg = fmt.Sprintf("%s", ClientStop)
		default:
			fmt.Fprintln(cmd.OutOrStderr(), "Invalid mode", mode)
			os.Exit(1)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Client sent:", msg)
		_, err = fmt.Fprintln(c, string([]byte(msg)))
		if err != nil {
			fmt.Fprintln(cmd.OutOrStderr(), "Write error:", err)
			os.Exit(1)
		}

		b := bufio.NewReader(c)

		line, err := b.ReadBytes('\n')
		if err != nil {
			fmt.Fprintln(cmd.OutOrStderr(), "Read error:", err)
			os.Exit(1)

		}
		fmt.Fprintln(cmd.OutOrStdout(), "Response:", string(line[0:len(line)-1]))
	},
}
