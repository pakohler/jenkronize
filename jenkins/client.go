package jenkins

import (
	"crypto/tls"
	"encoding/json"
	"github.com/cavaliercoder/grab"
	"github.com/pakohler/jenkronize/logging"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

type JenkinsAPIClient struct {
	http     *http.Client
	grab     *grab.Client
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
		grab: grab.NewClient(),
		log:  logging.GetLogger(),
	}
	j.grab.HTTPClient = j.http
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

func (j *JenkinsAPIClient) GetBaseUrl() string {
	return j.baseUrl
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
		err = newJenkinsError("Request to "+url+" failed", err)
		j.log.Error.Print(err.Error())
		return []byte{}, err
	}
	defer resp.Body.Close()
	j.log.Info.Print("attempting to read response from " + url)
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (j *JenkinsAPIClient) DownloadFile(urlPath string, destDir string) error {
	url := j.cleanUrl(urlPath)
	urlSplit := strings.Split(url, "/")
	fileName := urlSplit[len(urlSplit)-1]
	filePath := path.Join(destDir, fileName)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		os.MkdirAll(destDir, 0700)
	}
	j.log.Info.Print("Download starting: " + url)
	// since some artifacts are large and connections are unstable, we'll use
	// `grab` with auto-resume enabled for the actual download
	grabReq, _ := grab.NewRequest(filePath, url)
	grabReq.HTTPRequest.SetBasicAuth(j.user, j.password)
	resp := j.grab.Do(grabReq)
	<-resp.Done
	if err := resp.Err(); err != nil {
		err = newJenkinsError("Download failed: "+url, err)
		j.log.Error.Print(err.Error())
		return err
	}
	j.log.Info.Print("Download complete: " + url)
	return nil
}

func (j *JenkinsAPIClient) GetLastSuccessfulBuildForJob(jobPath string) (*Build, error) {
	j.log.Info.Print("Attempting to get the last successful build for " + jobPath)
	resp, err := j.getJson(jobPath)
	if err != nil {
		j.log.Error.Print(err.Error())
		return nil, err
	}
	var job Job
	err = json.Unmarshal(resp, &job)
	if err != nil {
		err = newJenkinsError(string(resp), err)
		j.log.Error.Print(err.Error())
		return nil, err
	}
	return job.LastSuccessfulBuild, nil
}

func (j *JenkinsAPIClient) GetArtifactUrlsFromBuild(buildPath string) ([]string, error) {
	j.log.Info.Print("attempting to get artifact URLs from " + buildPath)
	artifactUrls := []string{}
	resp, err := j.getJson(buildPath)
	if err != nil {
		j.log.Error.Print(err)
		return []string{}, err
	}
	var build JobBuild
	err = json.Unmarshal(resp, &build)
	if err != nil {
		err = newJenkinsError(string(resp), err)
		j.log.Error.Print(err.Error())
		return []string{}, err
	}
	for _, artifact := range build.Artifacts {
		url := build.Url + "artifact/" + artifact.RelativePath
		artifactUrls = append(artifactUrls, url)
	}
	return artifactUrls, nil
}
