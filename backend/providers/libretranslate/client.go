package libretranslate

import (
	"bytes"
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

type libretranslate struct {
	url    *url.URL
	client *http.Client
}

type translateReqBody struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type translateRespBody struct {
	TranslatedText string `json:"translatedText"`
}

func (w *libretranslate) Translate(inputLang string, outputLang string, text string) (string, error) {
	url := w.url.JoinPath(translateEndpoint)

	reqBody := translateReqBody{
		Q:      text,
		Source: inputLang,
		Target: outputLang,
	}
	reqBodyBytes, err := json.Marshal(&reqBody)
	if err != nil {
		return "", fmt.Errorf("cannot marshal request body: %w", err)
	}
	reqBodyReader := bytes.NewReader(reqBodyBytes)

	resp, err := w.client.Post(url.String(), "application/json", reqBodyReader)
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

	var parsedBody translateRespBody
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal response body: %w", err)
	}

	return parsedBody.TranslatedText, nil
}

func NewClient(path string) (*libretranslate, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	parsedUrl, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("cannot parse url %s: %w", path, err)
	}

	return &libretranslate{
		url:    parsedUrl,
		client: client,
	}, nil
}
