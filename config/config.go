package config

import (
	"github.com/go-yaml/yaml"
	"github.com/pakohler/Jenkronize/common"
	"github.com/pakohler/Jenkronize/logging"
	"github.com/pakohler/Jenkronize/tracking"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const configFileName = "config.yaml"

var config *Config

type JenkinsConfig struct {
	Username string
	Password string
	URL      string
}

type TrackerConfig struct {
	Interval    time.Duration
	TrackedJobs []*tracking.TrackedJob
}

type Config struct {
	Jenkins JenkinsConfig
	Tracker TrackerConfig
	LogFile string
	log     *logging.Logger
}

func (c *Config) setDefaultValues() *Config {
	c.log.Info.Print("Generating example config...")
	defaultDuration, err := time.ParseDuration("10m")
	if err != nil {
		logging.GetLogger().Fatal.Fatal(err)
	}
	c.Jenkins = JenkinsConfig{
		Username: "yourUserName",
		Password: "yourSecurePassword",
		URL:      "https://your.jenkins.fqdn/jenkins",
	}
	c.Tracker = TrackerConfig{
		Interval: defaultDuration,
		TrackedJobs: []*tracking.TrackedJob{
			tracking.NewTrackedJob("/job/SomeProject/job/Build/job/ABranchOrSomething", "/path/to/dir/to/cache/artifacts"),
		},
	}
	c.log = logging.GetLogger()
	dir := filepath.Dir(c.getFilePath())
	c.LogFile = filepath.Join(dir, "log.txt")
	return c
}

func (c *Config) getFilePath() string {
	dir, err := common.GetExeDir()
	if err != nil {
		c.log.Fatal.Fatal(err)
	}
	configPath := filepath.Join(dir, configFileName)
	return configPath
}

func load() *Config {
	c := &Config{}
	c.log = logging.GetLogger()
	configPath := c.getFilePath()
	configFile, err := os.Open(configPath)
	if err != nil {
		c.log.Fatal.Print(err)
		c.log.Fatal.Print("Config file does not exist or is unable to be opened: " + configPath)
		c.setDefaultValues()
		c.save()
		c.log.Fatal.Fatal("Please edit the config file at " + configPath + " before running again.")
	}
	defer configFile.Close()
	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		c.log.Fatal.Fatal(err)
	}
	err = yaml.Unmarshal(configBytes, c)
	if c.LogFile != "" {
		c.log.AddLogFile(c.LogFile)
	}
	c.log.Info.Print("Successfully loaded configuration from " + configPath)
	return c
}

func (c *Config) save() {
	configPath := c.getFilePath()
	c.log.Info.Print("Saving config to " + configPath)
	configFile, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		c.log.Fatal.Fatal(err)
	}
	defer configFile.Close()
	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		c.log.Fatal.Fatal(err)
	}
	configFile.Write(yamlBytes)
}

func Get() *Config {
	if config == nil {
		config = load()
	}
	return config
}
