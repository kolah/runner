package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"github.com/kolah/runner/config"
	"github.com/spf13/viper"
	"fmt"
	"github.com/kolah/runner/watcher"
	"os"
	"github.com/kolah/runner/server"
	"net"
	"bufio"
	"os/signal"
	"syscall"
	"strings"
	"runtime"
	"time"
	"github.com/kolah/runner/worker"
)

type ServerResponse string
type ClientCommand string

const (
	ServerOK  ServerResponse = "OK"
	ServerErr                = "ERR"
)

const (
	ClientSetMode ClientCommand = "SETMODE"
	ClientStop    ClientCommand = "STOP"
)

var (
	currentMode = worker.RunnerModeLiveRebuild
	runner      *worker.Runner
	builder     *worker.Builder
	watch       *watcher.Watcher
)

var rootCmd = &cobra.Command{
	Use:   "runner",
	Short: "Live reload tools",

	Run: func(cmd *cobra.Command, args []string) {

		os.Mkdir(config.Config.TmpPath, 0755)

		server := server.NewSocketServer(config.Config.RunnerPort, serverHandler)
		log.Println("Starting socket server")
		server.Start()

		watch = watcher.NewWatcher(config.Config.ProjectRoot, config.Config.TmpPath, config.Config.IgnoredDirectories, config.Config.ValidExtensions)
		watch.Start()

		builder = worker.NewBuilder(config.Config.ProjectRoot, config.Config.BuildPath(), config.Config.BuildLogPath())
		runner = worker.NewRunner(config.Config.BuildPath(), config.Config.DebuggerPort)

		loopIndex := 0
		buildFailed := false
		alreadyBuilt := false

		go func() {
			for {
				loopIndex++

				fmt.Printf("Waiting (loop %d)...\n", loopIndex)
				eventName := <-watch.EventChannel

				fmt.Printf("receiving first event %s\n", eventName)

				fmt.Printf("sleeping for %d milliseconds\n", config.Config.BuildDelay)
				time.Sleep(time.Duration(config.Config.BuildDelay) * time.Millisecond)

				fmt.Println("flushing events")
				flushFSEvents(watch)

				fmt.Printf("Started! (%d Goroutines)\n", runtime.NumGoroutine())

				if runner.GetMode() == worker.RunnerModeDebug {
					fmt.Println("ignoring code changes while debugging")
					continue
				}

				// extract filename from event
				fileName := strings.Replace(strings.Split(eventName, ":")[0], `"`, "", -1)
				if config.Config.HasFileValidExtension(fileName) || !alreadyBuilt {
					if err := builder.Build(); err != nil {
						buildFailed = true
						fmt.Println(err.Error())
					} else {
						buildFailed = false
						alreadyBuilt = true
					}
				}

				if !buildFailed {
					runner.SetMode(currentMode)
				} else if config.Config.WebWrapperEnabled {
					// start a web server to show the error
				}

				fmt.Printf(strings.Repeat("-", 20))
			}
		}()

		// trigger first build
		watch.EventChannel <- "/"

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

		sig := <-sigc
		log.Printf("Caught signal %s: shutting down.", sig)

		server.Stop()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func serverHandler(c net.Conn) {
	b := bufio.NewReader(c)

	for {
		line, err := b.ReadBytes('\n')
		if err != nil { // EOF, or worse
			break
		}
		// split command into parts, remove last character (
		parts := strings.Split(string(line[0:len(line)-1]), " ")

		switch ClientCommand(parts[0]) {
		case ClientStop:
			_, err := fmt.Fprintln(c, ServerOK)
			if err != nil {
				fmt.Println("Error returning response")
				c.Close()
				return
			}
			fmt.Println("Received STOP command")
			c.Close()
			os.Exit(0)
			break
		case ClientSetMode:
			if len(parts) != 2 {
				_, err := fmt.Fprintln(c, ServerErr, "Invalid number of arguments")
				if err != nil {
					fmt.Println("Error returning response")
					c.Close()
					return
				}
				fmt.Println("Invalid number of arguments")
				c.Close()
				return
			}
			mode := worker.RunnerMode(parts[1])
			if mode == worker.RunnerModeDebug || mode == worker.RunnerModeLiveRebuild {
				currentMode = mode
				if err := builder.Build(); err != nil {
					fmt.Fprintln(c, ServerErr, "Build error")
					c.Close()
					return
				}
				runner.SetMode(mode)
				fmt.Fprintln(c, ServerOK, "Switched mode to", mode)
				c.Close()
				return
			} else {
				fmt.Fprintln(c, ServerErr, "Unknown mode", mode)
			}
		default:
			fmt.Fprintln(c, ServerErr, "Unknown command", parts[0])
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(controlCmd)
}

func initConfig() {
	viper.SetConfigFile("runner.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("project_root", ".")
	viper.SetDefault("tmp_path", "./tmp")
	viper.SetDefault("build_filename", "runner-build")
	viper.SetDefault("build_log", "runner-build-errors.log")
	viper.SetDefault("valid_extensions", []string{".go"})
	viper.SetDefault("ignored_directories", []string{"tmp"})
	viper.SetDefault("build_delay", 600)
	viper.SetDefault("web_wrapper_enabled", false)
	viper.SetDefault("runner_port", 55555)
	viper.SetDefault("debugger_port", 2345)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.Unmarshal(&config.Config)
}

func flushFSEvents(w *watcher.Watcher) {
	for {
		select {
		case eventName := <-w.EventChannel:
			fmt.Printf("flushing event %s\n", eventName)
		default:
			return
		}
	}
}