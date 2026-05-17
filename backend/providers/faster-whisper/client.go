package fasterwhisper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	Name        = "faster_whisper"
	asrEndpoint = "/asr"
)

type fasterWhisper struct {
	url    *url.URL
	client *http.Client
}

type transcribeRespBody struct {
	Text string `json:"text"`
}

func (w *fasterWhisper) Transcribe(lang string, multipartHeader string, contentLength string, r io.Reader) (string, error) {
	url := w.url.JoinPath(asrEndpoint)

	q := url.Query()
	q.Add("language", lang)
	q.Add("output", "json")
	url.RawQuery = q.Encode()

	resp, err := w.client.Post(url.String(), multipartHeader, r)
	if err != nil {
		return "", fmt.Errorf("cannot send request: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("error status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var parsedBody transcribeRespBody
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal response body: %w", err)
	}

	return parsedBody.Text, nil
}

func NewClient(path string) (*fasterWhisper, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	parsedUrl, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("cannot parse url %s: %w", path, err)
	}

	return &fasterWhisper{
		url:    parsedUrl,
		client: client,
	}, nil
}
