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
	"path/filepath"
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
}

func (h *Tracker) Init() *Tracker {
	h.log = logging.GetLogger()
	h.trackedJobs = map[string]*TrackedJob{}
	h.notifiers = []notifications.Notifier{}
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
		if err != nil {
			h.notify(err.Error())
			h.log.Error.Print(err.Error())
			currentBuild = nil
		}
		if currentBuild == nil {
			// we'll wait the interval out and try again.
			time.Sleep(h.interval)
			continue
		}
		if currentBuild.Number > job.BuildNumber() {
			msg := fmt.Sprintf(
				"New build number %d for %s detected - last tracked was %d. Downloading artifacts...",
				currentBuild.Number,
				job.GetName(),
				job.BuildNumber(),
			)
			h.notify(msg)
			h.log.Info.Print(msg)
			err = h.handleNewBuild(job, currentBuild)
			// set and save the build state _after_ the artifacts are synced so they can be retried if something crashes
			if err != nil {
				msg = fmt.Sprintf(
					"Artifact download for tracked job %s's build number %d failed on one or more artifacts; will retry after wait interval.",
					job.GetName(),
					job.BuildNumber(),
				)
				h.log.Error.Print(msg)
				h.notify(msg)
			} else {
				job.SetBuild(currentBuild)
				msg = fmt.Sprintf(
					"Completed downloading artifacts for tracked job %s's build number %d.",
					job.GetName(),
					job.BuildNumber(),
				)
				h.notify(msg)
				h.log.Info.Print(msg)
				h.saveState()
			}
		} else {
			h.log.Info.Printf(
				"last observed build number %d for %s is up-to-date; no action required.",
				currentBuild.Number,
				job.GetName(),
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
		downloadChannels = append(downloadChannels, h.handleNewArtifact(job.GetName(), artifactUrl))
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

func (h *Tracker) handleNewArtifact(job string, url string) <-chan error {
	ch := make(chan error)
	go func() {
		err := h.client.DownloadFile(url, h.trackedJobs[job].SyncDir)
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
