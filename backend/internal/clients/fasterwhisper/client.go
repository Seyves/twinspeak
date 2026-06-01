package fasterwhisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

const (
	Name        = "faster_whisper"
	asrEndpoint = "/asr"
)

type Client struct {
	url    *url.URL
	client *http.Client
}

type TranscribeResp struct {
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
}

type Segment struct {
	Start float32 `json:"start"`
	End   float32 `json:"end"`
}

func (w *Client) Transcribe(ctx context.Context, lang string, in <-chan []byte) (transcription string, duration int, err error) {
	url := w.url.JoinPath(asrEndpoint)

	reqBody := &bytes.Buffer{}
	writer := multipart.NewWriter(reqBody)

	part, err := writer.CreateFormFile("audio_file", "file.wav")
	if err != nil {
		return "", 0, fmt.Errorf("cannot create form file: %w", err)
	}

	for data := range in {
		_, err := part.Write(data)
		if err != nil {
			return "", 0, fmt.Errorf("cannot write form file: %w", err)
		}
	}
	writer.Close()

	q := url.Query()
	q.Add("language", lang)
	q.Add("encode", "false")
	q.Add("output", "json")
	url.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := w.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("cannot send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		return "", 0, fmt.Errorf("cannot read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("error status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var parsedBody TranscribeResp
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return "", 0, fmt.Errorf("cannot unmarshal response body: %w", err)
	}

	if len(parsedBody.Segments) > 0 {
		lastIdx := len(parsedBody.Segments) - 1
		duration = int(parsedBody.Segments[lastIdx].End * 1000)
	}

	return parsedBody.Text, duration, nil
}

func NewClient(path string) (*Client, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	parsedUrl, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("cannot parse url %s: %w", path, err)
	}

	return &Client{
		url:    parsedUrl,
		client: client,
	}, nil
}
