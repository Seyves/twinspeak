package libretranslate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	Name              = "libretranslate"
	translateEndpoint = "/translate"
)

type Client struct {
	url    *url.URL
	client *http.Client
}

type TranslateReq struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type TranslateResp struct {
	TranslatedText string `json:"translatedText"`
}

func (w *Client) Translate(ctx context.Context, inputLang string, outputLang string, text string) (string, error) {
	url := w.url.JoinPath(translateEndpoint)

	reqBody := TranslateReq{
		Q:      text,
		Source: inputLang,
		Target: outputLang,
	}
	reqBodyBytes, err := json.Marshal(&reqBody)
	if err != nil {
		return "", fmt.Errorf("cannot marshal request body: %w", err)
	}
	reqBodyReader := bytes.NewReader(reqBodyBytes)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), reqBodyReader)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("error status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var parsedBody TranslateResp
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal response body: %w", err)
	}

	return parsedBody.TranslatedText, nil
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
