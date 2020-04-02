# Jenkins Artifact Mirroring Made Easy
## Because sometimes you need a local mirror of stuff

Jenkronize is a simple tool that leverages the Jenkins REST API to find artifacts for specified jobs, which it will then download into the specified local directory.
You could then, for instance, set up a simple HTTP fileserver make those artifacts available to local VMs, saving your internet bandwidth for other things.

## Coming Soon:
- asyncronous downloads
- slack notifications

## Running

Simply run `./Jenkronize` (on \*Nix systems) or `Jenkronize.exe` (on Windows).
If there is no preexisting `config.yaml` in the same dir as the executable, an example config will be generated.
State (eg. last observed build) is stored in `state.json`; currently, if you remove tracked jobs from your config you must also delete them from your state file (or simply delete the state file altogether, though this would mean that other tracked jobs will have their artifacts downloaded again)

## Configuration

### jenkins
- `username`: the username for accessing the Jenkins API via basic auth
- `password`: the password for accessing the Jenkins API via basic auth
- `url`: the full URL to your Jenkins instance, eg. `https://leeroy.jenkins.yourdomain.org` or `http://yourdomain.org/leeroy/jenkins`

### tracker
- `interval`: the time to wait between checks for new builds. It uses Go's `time.Duration` format, eg `10s`, `2m` or `1m13s24ns` - see https://golang.org/pkg/time/#ParseDuration
- `trackedjobs`: A list of Jenkins jobs you want to track and synchronize artifacts from. Each entry should include the following:
    - `name` should be the path after the Jenkins URL for the jobs you want to track; for example `/job/foo/job/bar`.
    - `sync_dir` is path to the directory where you want to cache the artifacts from that job.

### logfile
(optional) the path where you want to log output to. If omitted, logs will go to `stdout` and `stderr`

## Building

Builds are currently being done using go version 1.12.9

External dependencies are:
- github.com/go-yaml/yaml

You can install each of the external dependencies with eg. `go get github.com/go-yaml/yaml`

