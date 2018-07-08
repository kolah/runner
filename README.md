# runner

Runner is a command line tool that builds and (re)starts your web application every time you save a Go file file.
Additionally, it supports running a `dlv` debugger, which can be useful while running application in a Docker container.

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

## Configuration reference

```yaml
project_root: . # Path to project files to watch, defaults to current directory
tmp_path: "./tmp" # Temporary directory path
build_filename:  "runner-build" # Executable filename, relative to `tmp_path`
build_log: "runner-build-errors.log" # Build log filename, relative to `tmp_path`
valid_extensions: [".go"]
ignored_directories: ["tmp"]  
build_delay: 600 # Delay between builds in miliseconds
runner_port: 55555 # Port used to send commands to running application (`runner ctl`)
web_wrapper_enabled: false # Run web server on APPLICATION_PORT (default 80) to show build error
```
