package tracking

import (
	"github.com/pakohler/jenkronize/jenkins"
	"strings"
)

type TrackedJob struct {
	Name          string         `yaml:"name"`
	Alias         string         `yaml:"alias"`
	Build         *jenkins.Build `yaml:"-" json:"build"`
	SyncDir       string         `yaml:"sync_dir"`
	BuildsToCache int            `yaml:buildsToCache`
}

func NewTrackedJob(name string, alias string, syncDir string) *TrackedJob {
	name = strings.ToLower(name)
	name = strings.TrimRight(name, "/")
	t := TrackedJob{
		Name:          name,
		Alias:         alias,
		SyncDir:       syncDir,
		BuildsToCache: 1,
	}
	t.Init()
	return &t
}

func (t *TrackedJob) Init() {
	if t.Alias == "" {
		t.Alias = t.Name
	}
	if t.Build == nil {
		t.Build = &jenkins.Build{Number: 0}
	}
}

func (t *TrackedJob) SetBuild(new *jenkins.Build) *TrackedJob {
	t.Build = new
	return t
}

func (t *TrackedJob) GetBuild() *jenkins.Build {
	return t.Build
}

func (t *TrackedJob) BuildNumber() int32 {
	if t.Build == nil {
		t.Build = &jenkins.Build{Number: 0}
	}
	return t.Build.Number
}

func (t *TrackedJob) GetName() string {
	return t.Name
}

func (t *TrackedJob) GetAlias() string {
	return t.Alias
}

func (t *TrackedJob) Equals(other *TrackedJob) bool {
	// There shouldn't be any case where you end up with multiple instances of
	// the same job to be tracked, but if so, synchronize the last seen build
	// between them.
	if t.GetName() == other.GetName() {
		if t.BuildNumber() >= other.BuildNumber() {
			other.SetBuild(t.GetBuild())
		} else {
			t.SetBuild(other.GetBuild())
		}
		return true
	}
	return false
}
