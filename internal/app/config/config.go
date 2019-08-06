package config

import (
	"github.com/kolah/runner/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type Logging struct {
	Level string
}

type Watch struct {
	Directories        []string
	WatchPatterns      []string `mapstructure:"watch_patterns" yaml:"watch_patterns"`
	IgnoredDirectories []string `mapstructure:"ignored_directories" yaml:"ignored_directories"`
}

type Build struct {
	Delay    time.Duration
	Command  string
	ErrorLog string `mapstructure:"error_log" yaml:"error_log"`
	TmpDir   string `mapstructure:"tmp_dir" yaml:"tmp_dir"`
}

type Run struct {
	Command          string
	DebugCommand     string `mapstructure:"debug_command" yaml:"debug_command"`
	BuildBeforeDebug bool   `mapstructure:"build_before_debug" yaml:"build_before_debug"`
}

type Config struct {
	Watch   Watch
	Run     Run
	Build   Build
	Logging Logging
	CtlPort int `mapstructure:"ctl_port" yaml:"ctl_port"`
}

func LoadConfig(cmd *cobra.Command) (*Config, error) {
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("RUNNER")
	viper.AutomaticEnv()

	viper.SetDefault("ctl_port", 55555)

	viper.SetDefault("watch.directories", []string{"."})
	viper.SetDefault("watch.watch_patterns", []string{"*.go"})
	viper.SetDefault("watch.ignore_directories", []string{"tmp", "vendor"})

	viper.SetDefault("build.command", "go build -gcflags='all=-N=-l' -o tmp/tmp-build .")
	viper.SetDefault("build.error_log", "tmp/build_error.log")
	viper.SetDefault("build.delay", 650*time.Millisecond)
	viper.SetDefault("build.tmp_dir", "tmp")

	viper.SetDefault("run.command", "tmp/tmp-build")
	viper.SetDefault("run.debug_command", "dlv --headless --listen=:2345 --api-version=2 exec tmp/tmp-build")
	viper.SetDefault("run.build_before_debug", true)

	viper.SetDefault("logging.level", "info")

	if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigFile("runner.yaml")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigParseError:
			return nil, err
		}
	}
	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ConfigureLogging(config Logging) (app.Logger, error) {
	level, err := app.ParseLevel(config.Level)

	if err != nil {
		return nil, err
	}

	return app.NewStdoutLog(level), nil
}