package jenkins

import (
	"crypto/tls"
	"encoding/json"
	"github.com/pakohler/Jenkronize/logging"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

type JenkinsAPIClient struct {
	http     *http.Client
	baseUrl  string
	user     string
	password string
	log      *logging.Logger
}

func New() *JenkinsAPIClient {
	// bad to have this as a default, but fixing that is a TODO
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	j := JenkinsAPIClient{
		http: &http.Client{Transport: transport},
		log:  logging.GetLogger(),
	}
	return &j
}

func (j *JenkinsAPIClient) SetUser(user string) *JenkinsAPIClient {
	j.log.Info.Print("setting username to " + user)
	j.user = user
	return j
}

func (j *JenkinsAPIClient) SetPassword(pass string) *JenkinsAPIClient {
	j.log.Info.Print("set password")
	j.password = pass
	return j
}

func (j *JenkinsAPIClient) SetBaseUrl(baseUrl string) *JenkinsAPIClient {
	j.log.Info.Print("set base URL to " + baseUrl)
	j.baseUrl = strings.TrimRight(baseUrl, "/")
	return j
}

func (j *JenkinsAPIClient) cleanUrl(urlPath string) string {
	urlPath = strings.TrimRight(urlPath, "/")
	urlPath = strings.ReplaceAll(urlPath, j.baseUrl, "")
	url := j.baseUrl + urlPath
	return url
}

func (j *JenkinsAPIClient) getJson(urlPath string) ([]byte, error) {
	url := j.cleanUrl(urlPath) + "/api/json"
	j.log.Info.Print("GETing " + url)
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(j.user, j.password)
	resp, err := j.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	j.log.Info.Print("attempting to read response from " + url)
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (j *JenkinsAPIClient) DownloadFile(urlPath string, destDir string) {
	url := j.cleanUrl(urlPath)
	urlSplit := strings.Split(url, "/")
	fileName := urlSplit[len(urlSplit)-1]
	filePath := path.Join(destDir, fileName)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		os.MkdirAll(destDir, os.ModeDir)
	}
	out, err := os.Create(filePath)
	if err != nil {
		j.log.Fatal.Fatal(err)
	}
	defer out.Close()
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(j.user, j.password)
	resp, err := j.http.Do(req)
	if err != nil {
		j.log.Fatal.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		j.log.Error.Printf("Download of %s had a bad status: %s", url, resp.Status)
		return
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		j.log.Error.Printf("Download of %s failed: %v", url, err)
	}
}

func (j *JenkinsAPIClient) GetLastSuccessfulBuildForJob(jobPath string) *Build {
	j.log.Info.Print("Attempting to get the last successful build for " + jobPath)
	resp, err := j.getJson(jobPath)
	if err != nil {
		j.log.Fatal.Fatal(err)
	}
	var job Job
	err = json.Unmarshal(resp, &job)
	if err != nil {
		j.log.Fatal.Print(string(resp))
		j.log.Fatal.Fatal(err)
	}
	return job.LastSuccessfulBuild
}

func (j *JenkinsAPIClient) GetArtifactUrlsFromBuild(buildPath string) []string {
	j.log.Info.Print("attempting to get artifact URLs from " + buildPath)
	resp, err := j.getJson(buildPath)
	if err != nil {
		j.log.Fatal.Fatal(err)
	}
	var build JobBuild
	err = json.Unmarshal(resp, &build)
	if err != nil {
		j.log.Fatal.Print(string(resp))
		j.log.Fatal.Fatal(err)
	}
	var artifactUrls []string
	for _, artifact := range build.Artifacts {
		url := build.Url + "artifact/" + artifact.RelativePath
		artifactUrls = append(artifactUrls, url)
	}
	return artifactUrls
}
