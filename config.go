package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/yiblet/hlp/chat"
)

const defaultConfigFilename = "configuration.json"

type config struct {
	OpenAIAPIKey      string `json:"openai_api_key"`
	OpenAIAPIEndpoint string `json:"endpoint,omitempty"`
	DefaultModel      string `json:"model,omitempty"`
	fileName          string
	debug             bool
}

func (c *config) Model() string {
	if c.DefaultModel == "" {
		return "gpt-4o-mini"
	}
	return c.DefaultModel
}

type loggingRoundTripper struct{ inner http.RoundTripper }

func (l loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("Request: %s %s\n", req.Method, req.URL.String())
	for k, v := range req.Header {
		fmt.Printf("%s: %s\n", k, v)
	}

	if req.ContentLength > 0 && req.ContentLength < 1024*16 {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s\n", string(buf))
		req.Body = io.NopCloser(bytes.NewBuffer(buf))
	}
	fmt.Println()

	resp, err := l.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Response: %s\n", resp.Status)
	for k, v := range resp.Header {
		fmt.Printf("%s: %s\n", k, v)
	}

	if resp.ContentLength > 0 && resp.ContentLength < 1024*16 {
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s\n", string(buf))
		resp.Body = io.NopCloser(bytes.NewBuffer(buf))
	}
	fmt.Println()

	return resp, nil
}

func (c *config) Client() chat.Streamer {
	httpClient := &http.Client{
		Timeout: 0,
	}

	if c.debug {
		httpClient.Transport = loggingRoundTripper{inner: httpClient.Transport}
	}

	openai.NewClient()

	opts := []option.RequestOption{
		option.WithHTTPClient(httpClient),
	}

	if c.OpenAIAPIEndpoint != "" {
		opts = append(opts, option.WithBaseURL(c.OpenAIAPIEndpoint))
	}

	if c.OpenAIAPIKey != "" {
		opts = append(opts, option.WithAPIKey(c.OpenAIAPIKey))
	}

	client := openai.NewClient(opts...)

	return chat.NewOpenAIStreamer(client)
}

func getConfigPath() string {
	// A common use case is to get a private config folder for your app to
	// place its settings files into, that are specific to the local user.
	return configdir.LocalConfig("hlp")
}

func (c *config) Write() error {
	fileName := c.fileName
	if fileName == "" {
		fileName = defaultConfigFilename
	}

	configPath := getConfigPath()
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		return fmt.Errorf("cannot read path: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, fileName)
	fh, err := os.Create(configFile)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	encoder := json.NewEncoder(fh)
	return encoder.Encode(c)
}

func ReadConfig(fileName string, debug bool) (config, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = defaultConfigFilename
	}

	c := config{
		debug: debug,
	}
	// A common use case is to get a private config folder for your app to
	// place its settings files into, that are specific to the local user.
	configPath := getConfigPath()
	err := os.MkdirAll(configPath, 0755) // Ensure it exists.
	if err != nil {
		return config{}, fmt.Errorf("cannot read path: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, fileName)
	if _, err = os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			return config{
				fileName: fileName,
			}, nil
		}
		return config{}, err
	}

	// Load the existing file.
	fh, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	decoder := json.NewDecoder(fh)
	if err := decoder.Decode(&c); err != nil {
		return config{}, err
	}

	c.fileName = fileName
	return c, nil
}
