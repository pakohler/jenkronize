package main

import (
	"github.com/pakohler/jenkronize/config"
	"github.com/pakohler/jenkronize/jenkins"
	"github.com/pakohler/jenkronize/notifications"
	"github.com/pakohler/jenkronize/tracking"
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

	if conf.Slack.Webhook != "" {
		slack := notifications.NewSlackNotifier(conf.Slack.Webhook)
		if conf.Slack.Channel != "" {
			slack.SetChannel(conf.Slack.Channel)
		}
		tracker.AddNotifier(slack)
	}

	for _, job := range conf.Tracker.TrackedJobs {
		tracker.Track(job)
	}

	tracker.LoadState()
	tracker.Go()
}
