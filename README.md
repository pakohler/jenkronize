# Jenkins Artifact Mirroring Made Easy
## Because sometimes you need a local mirror of stuff

Jenkronize is a simple tool that leverages the Jenkins REST API to find artifacts for specified jobs, which it will then download into the specified local directory.
You could then, for instance, set up a simple HTTP fileserver make those artifacts available to local VMs, saving your internet bandwidth for other things.

## Running

Simply run `./jenkronize` (on \*Nix systems) or `jenkronize.exe` (on Windows).
If there is no preexisting `config.yaml` in the same dir as the executable, an example config will be generated.
State (eg. last observed build) is stored in `state.json`; currently, if you remove tracked jobs from your config you must also delete them from your state file (or simply delete the state file altogether, though this would mean that other tracked jobs will have their artifacts downloaded again)

## Configuration

An example configuration:
```yaml
jenkins:
  username: theJenkinsUser
  password: theJenkinsUsersPassword!VerySecure?
  url: https://some.domain.fqdn/leeroy/jenkins
tracker:
  interval: 10m0s
  trackedjobs:
  - name: /job/installer/job/master
    sync_dir: /opt/jenkins-sync/installer
  - name: /job/database-access-layer/job/master
    sync_dir: /opt/jenkins-sync/database-access-layer
  - name: /job/web-interface
    sync_dir: /opt/jenkins-sync/web-interface
slack:
  webhook: https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX
  channel: "#jenkronize"
logfile: /opt/jenkins-sync/jenkronize.log
```

### jenkins
- `username`: the username for accessing the Jenkins API via basic auth
- `password`: the password for accessing the Jenkins API via basic auth
- `url`: the full URL to your Jenkins instance, eg. `https://leeroy.jenkins.yourdomain.org` or `http://yourdomain.org/leeroy/jenkins`

### tracker
- `interval`: the time to wait between checks for new builds. It uses Go's `time.Duration` format, eg `10s`, `2m` or `1m13s24ns` - see https://golang.org/pkg/time/#ParseDuration
- `trackedjobs`: A list of Jenkins jobs you want to track and synchronize artifacts from. Each entry should include the following:
    - `name` should be the path after the Jenkins URL for the jobs you want to track; for example `/job/foo/job/bar`.
    - `sync_dir` is path to the directory where you want to cache the artifacts from that job. If the dir doesn't exist, Jenkronize will attempt to create it.

### slack
- `webhook`: (optional) an incoming webhook for Slack notifications.
- `channel`: (optional) the Slack channel to post notifications.

### logfile
(optional) the path where you want to log output to. If omitted, logs will go to `stdout` and `stderr`

## Building

Builds are currently using Go version 1.12.9

External dependencies include:
- `github.com/go-yaml/yaml`
- `github.com/cavaliercoder/grab`
They can be installed with eg. `go get github.com/go-yaml/yaml`

Once the dependencies are installed, simply `cd` into the top-level directory of the repository and `go build .`
