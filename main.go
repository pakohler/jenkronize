package main

import (
	"github.com/pakohler/Jenkronize/config"
	"github.com/pakohler/Jenkronize/jenkins"
	"github.com/pakohler/Jenkronize/tracking"
)

func main() {
	conf := config.Get()

	leeroy := jenkins.
		New().
		SetUser(conf.Jenkins.Username).
		SetPassword(conf.Jenkins.Password).
		SetBaseUrl(conf.Jenkins.URL)

	tracker := (&tracking.Tracker{}).
		Init().
		SetClient(leeroy).
		SetInterval(conf.Tracker.Interval.String())

	for _, job := range conf.Tracker.TrackedJobs {
		tracker.Track(job)
	}

	tracker.LoadState()
	tracker.Go()
}
