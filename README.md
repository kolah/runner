# runner

Runner is a command line tool that builds and (re)starts your web application every time you save a Go file file.
It's similar to live rebuild tools available for JS.
 
It supports two modes of operation: `live rebuild` and `debug`. Debug might be useful for running a debugger, which can be useful while running application in a Docker container.

The tool was intended to be used for Go projects, but in current state it's language agnostic – 
you should be able to tweak it to your needs.   

Initially it was based on [fresh](https://github.com/pilu/fresh).

## Installation

    go get github.com/kolah/runner

## Usage

    cd /path/to/myapp

Start runner:

    runner

Runner will watch for file events, and every time you create/modify/delete a file it will build and restart the application.
If `go build` returns an error, it will log it in the tmp folder.

### Switching modes

```bash
runner ctl debug # switch to debug mode
runner ctl rebuild # switch to rebuild mode
runner ctl stop # terminates runner
``` 

## Configuration
Runner looks for a `runner.yaml` configuration file in current directory. For a list of options, see the configuration reference below. 

You can also set environment variables. `runner` tries to read them using `RUNNER_` prefix. 
In order to set `watch.ignored_directories` you can set `RUNNER_WATCH_IGNORED_DIRECTORIES=/dir1,/dir2`. Environment variables overwrite the configuration file settings.

If a configuration option is not set, `runner` falls back to default values.

## Configuration reference

Reference below contains all available options with the default values.

```yaml
ctl_port: 55555 # Listen on this port to enable changing state of running instance
watch:
    directories: # A list of directories to watch
        - .
    watch_patterns: # A list of file patterns to trigger rebuild. You can use wildcards or enter exact filenames
        - "*.go"
    ignored_directories: ["tmp", "vendor"] # A list of directories not to watch
    verbose: false
build:
    command: go build -gcflags='all=-N -l' -o tmp/tmp-build . # Command triggered to build the application
    error_log: tmp/build_error.log # Location of the build error log file.
    delay: 650ms # Delay before build that is triggered by file system changes
    tmp_dir: tmp # Location of tmp dir. It will be created recursively on start if not exists
run:
    command: tmp/tmp-build
    debug_command: dlv --headless --listen=:2345 --api-version=2 exec tmp/tmp-build # Command triggered to start debug
    build_before_debug: true # Flag executing build before debug

```
