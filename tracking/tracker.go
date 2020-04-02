package tracking

import (
	"encoding/json"
	"github.com/pakohler/Jenkronize/common"
	"github.com/pakohler/Jenkronize/jenkins"
	"github.com/pakohler/Jenkronize/logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Tracker struct {
	client      *jenkins.JenkinsAPIClient
	log         *logging.Logger
	trackedJobs map[string]*TrackedJob
	interval    time.Duration
}

func (h *Tracker) Init() *Tracker {
	h.log = logging.GetLogger()
	h.trackedJobs = map[string]*TrackedJob{}
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

func (h *Tracker) Go() {
	for _, trackedJob := range h.trackedJobs {
		go h.TrackJob(trackedJob)
	}
	for {
		time.Sleep(10000)
	}
}

func (h *Tracker) TrackJob(job *TrackedJob) {
	for {
		currentBuild := h.client.GetLastSuccessfulBuildForJob(job.GetName())
		if currentBuild.Number > job.BuildNumber() {
			h.log.Info.Printf("new build number %d for %s detected; last tracked was %d", currentBuild.Number, job.GetName(), job.BuildNumber())
			h.handleNewBuild(job, currentBuild)
			// set and save the build state _after_ the artifacts are synced so they can be retried if something crashes
			job.SetBuild(currentBuild)
			h.saveState()
		} else {
			h.log.Info.Printf("last observed build number %d for %s is up-to-date; no action required.", currentBuild.Number, job.GetName())
		}
		time.Sleep(h.interval)
	}
}

func (h *Tracker) handleNewBuild(job *TrackedJob, newBuild *jenkins.Build) {
	artifacts := h.client.GetArtifactUrlsFromBuild(newBuild.Url)
	// kick off all the downloads; when they're complete, their channel will recieve a `true` bool
	downloadChannels := make([]<-chan bool, 0)
	for _, artifactUrl := range artifacts {
		downloadChannels = append(downloadChannels, h.handleNewArtifact(job.GetName(), artifactUrl))
	}
	// wait for all downloads to complete
	for _, c := range downloadChannels {
		<-c
	}
}

func (h *Tracker) handleNewArtifact(job string, url string) <-chan bool {
	ch := make(chan bool)
	h.log.Info.Printf("%s artifact: %s", job, url)
	go func() {
		h.client.DownloadFile(url, h.trackedJobs[job].SyncDir)
		ch <- true
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
