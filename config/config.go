package config

import (
	_ "gopkg.in/yaml.v2"
	"path/filepath"
	"runtime"
	"strings"
)

var Config configuration

type configuration struct {
	ProjectRoot         string   `yaml:"project_root"`
	TmpPath             string   `mapstructure:"tmp_path" yaml:"tmp_path"`
	BuildFilename       string   `mapstructure:"build_filename" yaml:"build_filename"`
	BuildLog            string   `mapstructure:"build_log" yaml:"build_log"`
	ValidExtensions     []string `mapstructure:"valid_extensions" yaml:"valid_extensions"`
	IgnoredDirectories  []string `mapstructure:"ignored_directories" yaml:"ignored_directories"`
	BuildDelay          int      `mapstructure:"build_delay" yaml:"build_delay"`
	WebWrapperEnabled   bool     `mapstructure:"web_wrapper_enabled" yaml:"web_wrapper_enabled"`
	RunnerPort          int      `mapstructure:"runner_port" yaml:"runner_port"`
	DebuggerPort        int      `mapstructure:"debugger_port" yaml:"debugger_port"`
}

func (c configuration) BuildLogPath() string {
	return filepath.Join(c.TmpPath, c.BuildLog)
}

func (c configuration) BuildPath() string {
	p := filepath.Join(c.TmpPath, c.BuildFilename)
	if runtime.GOOS == "windows" && filepath.Ext(p) != ".exe" {
		p += ".exe"
	}
	return p
}

func (c configuration) HasFileValidExtension(fileName string) bool {
	for _, e := range c.ValidExtensions {
		e = strings.TrimSpace(e)
		if strings.HasSuffix(fileName, e) {
			return true
		}
	}

	return false
}

func init() {
	Config = configuration{}
}
