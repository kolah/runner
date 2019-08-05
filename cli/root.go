package cli

import (
	"github.com/kolah/runner/internal/app"
	"github.com/kolah/runner/internal/app/config"
	"github.com/kolah/runner/internal/app/rpc"
	"github.com/kolah/runner/internal/pkg/simplerpc"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var rootCmd = cobra.Command{
	Use:   "runner",
	Short: "Live reload tools",
	Run:   run,
}

func RootCommand() *cobra.Command {
	rootCmd.PersistentFlags().StringP("config", "c", "", "the config file to use")
	rootCmd.AddCommand(controlCmd)

	return &rootCmd
}

func run(cmd *cobra.Command, args []string) {
	configuration, err := config.LoadConfig(cmd)
	if err != nil {
		log.Fatal("Failed to load config: " + err.Error())
	}

	_ = os.MkdirAll(configuration.Build.TmpDir, 0755)

	builder := app.NewBuilder(configuration.Build.Command, configuration.Build.ErrorLog)
	watch := app.NewWatcher(configuration.Watch.Directories, configuration.Watch.IgnoredDirectories, configuration.Watch.WatchPatterns)

	runnerOptions := app.NewRunnerOptions(configuration.Build.Delay, configuration.Run.Command, configuration.Run.DebugCommand, configuration.Run.BuildBeforeDebug)
	runner := app.NewRunner(watch, builder, runnerOptions)

	server := simplerpc.NewServer(configuration.CtlPort)
	server.AddHandler(string(rpc.ClientStop), rpc.StopHandler())
	server.AddHandler(string(rpc.ClientSetMode), rpc.SetModeHandler(runner))

	log.Printf("Starting TCP server for commands on port %d", configuration.CtlPort)
	server.Start()

	if err := runner.Start(); err != nil {
		log.Fatalf("Failed to start runner: %s", err.Error())
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	sig := <-sigc
	log.Printf("Caught signal %s: shutting down.", sig)

	if err := runner.Stop(); err != nil {
		log.Printf("Failed to stop runner: %s" + err.Error())
	}
	//noinspection ALL
	server.Stop()
}
