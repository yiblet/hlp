package main

import (
	"encoding/json"
	"fmt"
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
	OpenaiAPIEndpoint string `json:"endpoint,omitempty"`
	DefaultModel      string `json:"model,omitempty"`
	fileName          string
}

func (c *config) Model() string {
	if c.DefaultModel == "" {
		return "gpt-4o-mini"
	}
	return c.DefaultModel
}

func (c *config) Client() chat.Streamer {
	httpClient := &http.Client{
		Timeout: 0,
	}

	openai.NewClient()

	opts := []option.RequestOption{
		option.WithHTTPClient(httpClient),
	}

	if c.OpenaiAPIEndpoint != "" {
		opts = append(opts, option.WithBaseURL(c.OpenaiAPIEndpoint))
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

func ReadConfig(fileName string) (config, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = defaultConfigFilename
	}

	c := config{}
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
