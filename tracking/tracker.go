package tracking

import (
	"encoding/json"
	"fmt"
	"github.com/pakohler/jenkronize/common"
	"github.com/pakohler/jenkronize/jenkins"
	"github.com/pakohler/jenkronize/logging"
	"github.com/pakohler/jenkronize/notifications"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type comboError struct {
	errorSet []error
}

func (c *comboError) Error() string {
	str := "Multi-Error combination:\n"
	for _, e := range c.errorSet {
		str += e.Error() + "\n"
	}
	return str
}

type Tracker struct {
	client      *jenkins.JenkinsAPIClient
	log         *logging.Logger
	trackedJobs map[string]*TrackedJob
	interval    time.Duration
	notifiers   []notifications.Notifier
	mux         sync.Mutex
	dns         bool
}

func (h *Tracker) Init() *Tracker {
	h.log = logging.GetLogger()
	h.trackedJobs = map[string]*TrackedJob{}
	h.notifiers = []notifications.Notifier{}
	h.dns = true
	return h
}

func (h *Tracker) SetClient(new *jenkins.JenkinsAPIClient) *Tracker {
	h.client = new
	return h
}

func (h *Tracker) SetInterval(new string) *Tracker {
	duration, err := time.ParseDuration(new)
	if err != nil {
		h.log.Fatal.Fatal(err)
	}
	h.interval = duration
	return h
}

func (h *Tracker) Track(job *TrackedJob) *Tracker {
	_, ok := h.trackedJobs[job.GetName()]
	if !ok {
		h.trackedJobs[job.GetName()] = job
	}
	return h
}

func (h *Tracker) AddNotifier(newNotifier notifications.Notifier) *Tracker {
	h.notifiers = append(h.notifiers, newNotifier)
	return h
}

func (h *Tracker) Go() {
	for _, trackedJob := range h.trackedJobs {
		go h.TrackJob(trackedJob)
	}
	for {
		time.Sleep(10000)
	}
}

func (h *Tracker) notify(msg string) {
	for _, n := range h.notifiers {
		err := n.Post(msg)
		if err != nil {
			h.log.Error.Print(err.Error())
		}
	}
}

func (h *Tracker) TrackJob(job *TrackedJob) {
	for {
		currentBuild, err := h.client.GetLastSuccessfulBuildForJob(job.GetName())
		h.mux.Lock()
		if err != nil {
			h.log.Error.Print(err.Error())
			if strings.Contains(err.Error(), "dial tcp: lookup") {
				// special handling for common DNS issues
				if h.dns {
					// We'll only send notifications when we used to be able to reach the host,
					// but can't now, to avoid being too spammy.
					h.notify(fmt.Sprintf(
						"DNS lookup failed for Jenkins server %s - check your VPN, DNS, or network connectivity",
						h.client.GetBaseUrl(),
					))
				}
				h.dns = false
			} else {
				// send notifications of the error message
				h.notify(err.Error())
			}
			// we'll wait the interval out and try again.
			h.mux.Unlock()
			time.Sleep(h.interval)
			continue
		}
		// if we got here, we know we can reach the host.
		h.dns = true
		h.mux.Unlock()
		if currentBuild.Number > job.BuildNumber() {
			msg := fmt.Sprintf(
				"%s - new build number %d detected - last tracked was %d. Downloading artifacts...",
				job.GetAlias(),
				currentBuild.Number,
				job.BuildNumber(),
			)
			h.notify(msg)
			h.log.Info.Print(msg)
			err = h.handleNewBuild(job, currentBuild)
			// set and save the build state _after_ the artifacts are synced so they can be retried if something crashes
			if err != nil {
				msg = fmt.Sprintf(
					"%s - artifact download for build number %d failed on one or more artifacts; will retry after wait interval.",
					job.GetAlias(),
					job.BuildNumber(),
				)
				h.log.Error.Print(msg)
				h.notify(msg)
			} else {
				job.SetBuild(currentBuild)
				msg = fmt.Sprintf(
					"%s - completed downloading artifacts for build number %d.",
					job.GetAlias(),
					job.BuildNumber(),
				)
				h.notify(msg)
				h.log.Info.Print(msg)
				h.saveState()
			}
		} else {
			h.log.Info.Printf(
				"%s - last observed build number %d is up-to-date; no action required.",
				job.GetAlias(),
				currentBuild.Number,
			)
		}
		time.Sleep(h.interval)
	}
}

func (h *Tracker) handleNewBuild(job *TrackedJob, newBuild *jenkins.Build) error {
	artifacts, err := h.client.GetArtifactUrlsFromBuild(newBuild.Url)
	if err != nil {
		h.notify(err.Error())
		h.log.Error.Print(err.Error())
		return err
	}
	// kick off all the downloads; when they're complete, their channel will recieve an error
	// or `nil` if the download was successful
	downloadChannels := make([]<-chan error, 0)
	for _, artifactUrl := range artifacts {
		downloadChannels = append(downloadChannels, h.handleNewArtifact(job.GetName(), newBuild.Number, artifactUrl))
	}
	errorSet := []error{}
	// wait for all downloads to complete
	for _, c := range downloadChannels {
		err := <-c
		if err != nil {
			errorSet = append(errorSet, err)
			h.notify(err.Error())
			h.log.Error.Print(err.Error())
		}
	}
	if len(errorSet) > 0 {
		return &comboError{errorSet: errorSet}
	}
	return nil
}

func (h *Tracker) handleNewArtifact(job string, build int32, url string) <-chan error {
	ch := make(chan error)
	downloadDir := path.Join(h.trackedJobs[job].SyncDir, fmt.Sprintf("%d", build))
	go func() {
		err := h.client.DownloadFile(url, downloadDir)
		ch <- err
	}()
	return ch
}

func (h *Tracker) getStateFile() (*os.File, error) {
	dir, err := common.GetExeDir()
	if err != nil {
		h.log.Fatal.Fatal(err)
	}
	stateFilePath := filepath.Join(dir, "state.json")
	return os.OpenFile(stateFilePath, os.O_RDWR|os.O_CREATE, 0600)
}

func (h *Tracker) saveState() {
	file, err := h.getStateFile()
	if err != nil {
		h.log.Error.Printf("unable to open state file for saving: %v", err)
		return
	}
	defer file.Close()

	stateBytes, err := json.Marshal(h.trackedJobs)
	if err != nil {
		h.log.Error.Printf("unable to marshal state for saving: %v", err)
		return
	}
	file.Write(stateBytes)
}

func (h *Tracker) LoadState() {
	file, err := h.getStateFile()
	if err != nil {
		h.log.Error.Printf("unable to open state file for loading: %v", err)
		return
	}
	defer file.Close()
	tmpJobs := map[string]*TrackedJob{}
	stateBytes, err := ioutil.ReadAll(file)
	err = json.Unmarshal(stateBytes, &tmpJobs)
	if err != nil {
		h.log.Error.Printf("unable to load state from file: %v", err)
		return
	}
	for key, val := range tmpJobs {
		h.trackedJobs[key] = val
	}
}
