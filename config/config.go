package config

import (
	_ "gopkg.in/yaml.v2"
	"path/filepath"
	"runtime"
	"strings"
)

var Config configuration

type configuration struct {
	Root                string   `yaml:"root"`
	TmpPath             string   `mapstructure:"tmp_path" yaml:"tmp_path"`
	BuildName           string   `mapstructure:"build_name" yaml:"build_name"`
	BuildLog            string   `mapstructure:"build_log" yaml:"build_log"`
	ValidExtensions     []string `mapstructure:"valid_extensions" yaml:"valid_extensions"`
	NoRebuildExtensions []string `mapstructure:"no_rebuild_extensions" yaml:"no_rebuild_extensions"`
	IgnoredDirectories  []string `mapstructure:"ignored_directories" yaml:"ignored_directories"`
	BuildDelay          int      `mapstructure:"build_delay" yaml:"build_delay"`
	WebWrapperEnabled   bool     `mapstructure:"web_wrapper_enabled" yaml:"web_wrapper_enabled"`
	SocketPath          string   `mapstructure:"socket_path" yaml:"socket_path"`
}

func (c configuration) BuildLogPath() string {
	return filepath.Join(c.TmpPath, c.BuildLog)
}

func (c configuration) BuildPath() string {
	p := filepath.Join(c.TmpPath, c.BuildName)
	if runtime.GOOS == "windows" && filepath.Ext(p) != ".exe" {
		p += ".exe"
	}
	return p
}

func (c configuration) HasFileValidExtension(fileName string) bool {
	for _, e := range c.ValidExtensions {
		e = strings.TrimSpace(e)
		if strings.HasSuffix(fileName, e) {
			return false
		}
	}

	return true
}


func init() {
	Config = configuration{}
}
