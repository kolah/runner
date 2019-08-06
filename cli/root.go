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

	logger, err := config.ConfigureLogging(configuration.Logging)
	if err != nil {
		log.Fatal("Failed to configure logger: ", err.Error())
	}

	// colored output for running application
	appLogger := app.NewAppLog(logger)

	builder := app.NewBuilder(configuration.Build.Command, configuration.Build.ErrorLog, logger)
	watch := app.NewWatcher(configuration.Watch.Directories, configuration.Watch.IgnoredDirectories, configuration.Watch.WatchPatterns, logger)

	runnerOptions := app.NewRunnerOptions(configuration.Build.Delay, configuration.Run.Command, configuration.Run.DebugCommand, configuration.Run.BuildBeforeDebug)
	runner := app.NewRunner(watch, builder, runnerOptions, logger, appLogger)

	server := simplerpc.NewServer(configuration.CtlPort)
	server.AddHandler(rpc.Stop, rpc.StopHandler())
	server.AddHandler(rpc.SetMode, rpc.SetModeHandler(runner))

	logger.Infof("Starting TCP server for commands on port %d\n", configuration.CtlPort)
	if err := server.Start(); err != nil {
		logger.Info("Failed to start TCP server\n")
		os.Exit(1)
	}

	if err := runner.Start(); err != nil {
		logger.Infof("Fatal error while starting runner %s\n", err.Error())
		os.Exit(1)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	sig := <-sigc

	logger.Debugf("Caught signal %s: shutting down\n", sig)

	if err := runner.Stop(); err != nil {
		logger.Debugf("Failed to stop runner: %s\n", err.Error())
	}
	//noinspection ALL
	server.Stop()
}
