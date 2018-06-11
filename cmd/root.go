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

type RunnerMode string
type ServerResponse string
type ClientCommand string

const (
	RunnerModeLiveRebuild RunnerMode = "LIVEREBUILD"
	RunnerModeDebug       RunnerMode = "DEBUG"
)

const (
	ServerOK  ServerResponse = "OK"
	ServerErr                = "ERR"
)

const (
	ClientSetMode ClientCommand = "SETMODE"
	ClientStop    ClientCommand = "STOP"
)

var (
	currentMode = RunnerModeLiveRebuild
	runner      *worker.Worker
	builder     *worker.Builder
)

var rootCmd = &cobra.Command{
	Use:   "runner",
	Short: "Live reload tools",

	Run: func(cmd *cobra.Command, args []string) {

		os.Mkdir(config.Config.TmpPath, 0755)
		removeBuildErrorsLog()

		server := server.NewSocketServer(config.Config.SocketPath, serverHandler)
		log.Println("Starting socket server")
		server.Start()

		w := watcher.NewWatcher(config.Config.Root, config.Config.TmpPath, config.Config.IgnoredDirectories, config.Config.ValidExtensions)
		w.Start()

		builder = worker.NewBuilder(config.Config.Root, config.Config.BuildPath())

		loopIndex := 0
		buildFailed := false

		go func() {
			for {
				loopIndex++

				fmt.Printf("Waiting (loop %d)...\n", loopIndex)
				eventName := <-w.EventChannel

				fmt.Printf("receiving first event %s\n", eventName)
				fmt.Printf("sleeping for %d milliseconds\n", config.Config.BuildDelay)

				time.Sleep(time.Duration(config.Config.BuildDelay) * time.Millisecond)

				fmt.Println("flushing events")

				flushEvents(w)

				fmt.Printf("Started! (%d Goroutines)", runtime.NumGoroutine())
				removeBuildErrorsLog()

				// extract filename from event
				fileName := strings.Replace(strings.Split(eventName, ":")[0], `"`, "", -1)
				if config.Config.HasFileValidExtension(fileName) {
					if err := builder.Build(); err != nil {
						buildFailed = true
						fmt.Printf("Build Failed: \n %s\n", err.Error())
						createBuildErrorsLog(err.Error())
					}
				}

				if runner != nil {
					runner.Stop()
				}

				if !buildFailed {
					switch currentMode {
					case RunnerModeLiveRebuild:
						runner = worker.NewWorker(config.Config.BuildPath())
						runner.Run()
						break
					case RunnerModeDebug:
						runner = worker.NewWorker("dlv", "--headless", "--listen=:2345", "--api-version=2", "exec", config.Config.BuildPath())
						runner.Run()

						go func() {
							<-runner.FinishedChannel
							// return to live rebuild
							currentMode = RunnerModeLiveRebuild
							w.EventChannel <- "/"
						}()

						break
					}
				} else if config.Config.WebWrapperEnabled {
					// start web server to show the error
				}

				fmt.Printf(strings.Repeat("-", 20))
			}
		}()

		w.EventChannel <- "/"

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
			mode := RunnerMode(parts[1])
			if mode == RunnerModeDebug || mode == RunnerModeLiveRebuild {
				currentMode = mode
				fmt.Fprintln(c, ServerOK, "Switched mode to", mode)
				fmt.Println("Switched mode to", mode)
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

	viper.SetDefault("root", ".")
	viper.SetDefault("tmp_path", "./tmp")
	viper.SetDefault("build_name", "runner-build")
	viper.SetDefault("build_log", "runner-build-errors.log")
	viper.SetDefault("valid_extensions", []string{".go", ".tpl", ".tmpl", ".html"})
	viper.SetDefault("no_rebuild_extensions", []string{".tpl", ".tmpl", ".html"})
	viper.SetDefault("ignored_directories", []string{"assets", "tmp"})
	viper.SetDefault("build_delay", 600)
	viper.SetDefault("web_wrapper_enabled", false)
	viper.SetDefault("socket_path", "./tmp/runner.sock") //

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.Unmarshal(&config.Config)
}

func flushEvents(w *watcher.Watcher) {
	for {
		select {
		case eventName := <-w.EventChannel:
			fmt.Printf("flushing event %s\n", eventName)
		default:
			return
		}
	}
}

func createBuildErrorsLog(message string) bool {
	file, err := os.Create(config.Config.BuildLogPath())
	if err != nil {
		return false
	}

	_, err = file.WriteString(message)
	if err != nil {
		return false
	}

	return true
}

func removeBuildErrorsLog() {
	if _, err := os.Stat(config.Config.BuildLogPath()); os.IsExist(err) {
		os.Remove(config.Config.BuildLogPath())
	}
}
