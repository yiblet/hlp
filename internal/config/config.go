package config

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
	"github.com/yiblet/hlp/internal/llm"
)

const defaultConfigFilename = "configuration.json"

type Config struct {
	OpenAIAPIKey      string `json:"openai_api_key"`
	OpenAIAPIEndpoint string `json:"endpoint,omitempty"`
	DefaultModel      string `json:"model,omitempty"`
	fileName          string
	debug             bool
}

func (c *Config) Model() string {
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

func (c *Config) Client() llm.Streamer {
	httpClient := &http.Client{Timeout: 0}
	if c.debug {
		httpClient.Transport = loggingRoundTripper{inner: httpClient.Transport}
	}

	openai.NewClient()

	opts := []option.RequestOption{option.WithHTTPClient(httpClient)}
	if c.OpenAIAPIEndpoint != "" {
		opts = append(opts, option.WithBaseURL(c.OpenAIAPIEndpoint))
	}
	if c.OpenAIAPIKey != "" {
		opts = append(opts, option.WithAPIKey(c.OpenAIAPIKey))
	}

	client := openai.NewClient(opts...)
	return llm.NewOpenAIStreamer(client)
}

func Path() string {
	return configdir.LocalConfig("hlp")
}

func (c *Config) Write() error {
	fileName := c.fileName
	if fileName == "" {
		fileName = defaultConfigFilename
	}

	configPath := Path()
	if err := configdir.MakePath(configPath); err != nil {
		return fmt.Errorf("cannot read path: %w", err)
	}

	configFile := filepath.Join(configPath, fileName)
	fh, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	encoder := json.NewEncoder(fh)
	return encoder.Encode(c)
}

func Read(fileName string, debug bool) (Config, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = defaultConfigFilename
	}

	c := Config{debug: debug}
	configPath := Path()
	err := os.MkdirAll(configPath, 0755)
	if err != nil {
		return Config{}, fmt.Errorf("cannot read path: %w", err)
	}

	configFile := filepath.Join(configPath, fileName)
	if _, err = os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			return Config{fileName: fileName}, nil
		}
		return Config{}, err
	}

	fh, err := os.Open(configFile)
	if err != nil {
		return Config{}, err
	}
	defer fh.Close()

	decoder := json.NewDecoder(fh)
	if err := decoder.Decode(&c); err != nil {
		return Config{}, err
	}

	c.fileName = fileName
	return c, nil
}
