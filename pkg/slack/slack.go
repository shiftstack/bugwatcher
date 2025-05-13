package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	httpClient *http.Client
}

func New() Client {
	return Client{httpClient: &http.Client{}}
}

func (c Client) Send(slackHook string, text string) error {
	var msg bytes.Buffer
	err := json.NewEncoder(&msg).Encode(struct {
		LinkNames bool   `json:"link_names"`
		Text      string `json:"text"`
	}{
		LinkNames: true,
		Text:      text,
	})
	if err != nil {
		return fmt.Errorf("error marshalling the message payload: %w", err)
	}

	res, err := c.httpClient.Post(
		slackHook,
		"application/JSON",
		&msg,
	)
	if err != nil {
		return fmt.Errorf("error sending the message: %w", err)
	}

	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusAccepted:
	default:
		return fmt.Errorf("unexpected status code %q sending the message:", res.Status)
	}

	return nil
}

func Link(text, url string) string {
	return "<" + text + "|" + url + ">"
}
