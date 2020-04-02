package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Slack struct {
	client  *http.Client
	webhook string
	channel string
}

func NewSlackNotifier(webHook string) *Slack {
	s := &Slack{
		client:  &http.Client{},
		webhook: webHook,
	}
	return s
}

func (s *Slack) SetWebhook(newHook string) *Slack {
	s.webhook = newHook
	return s
}

func (s *Slack) SetChannel(newChannel string) *Slack {
	s.channel = newChannel
	return s
}

func (s *Slack) Post(msg string) error {
	// satisfy the Notifier interface and post to Slack
	if s.webhook == "" {
		return fmt.Errorf("Slack notification impossible; no webhook specified")
	}
	jsonMap := map[string]interface{}{
		"text": msg,
	}
	if s.channel != "" {
		jsonMap["channel"] = s.channel
	}
	data, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data)
	resp, err := s.client.Post(s.webhook, "application/json", reader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		response, _ := ioutil.ReadAll(resp.Body)
		responseText := string(response)
		return fmt.Errorf(
			"Slack notification POST reseponse status was %s; response text was: %s",
			resp.Status,
			responseText,
		)
	}
	return nil
}
