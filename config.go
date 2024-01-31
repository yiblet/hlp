package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/kirsle/configdir"
)

type config struct {
	OpenAIAPIKey string `json:"openai_api_key"`
	DefaultModel string `json:"model,omitempty"`
}

func (c *config) Model() string {
	if c.DefaultModel == "" {
		return "gpt-3.5-turbo"
	}
	return c.DefaultModel
}

func (c *config) Client() gpt3.Client {
	httpClient := &http.Client{
		Timeout: 0,
	}
	client := gpt3.NewClient(
		c.OpenAIAPIKey,
		gpt3.WithHTTPClient(httpClient),
		gpt3.WithDefaultEngine(c.Model()),
	)
	return client
}

func WriteConfig(c *config) error {
	// A common use case is to get a private config folder for your app to
	// place its settings files into, that are specific to the local user.
	configPath := configdir.LocalConfig("hlp")
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		return fmt.Errorf("cannot read path: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, "configuration.json")
	fh, err := os.Create(configFile)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	encoder := json.NewEncoder(fh)
	return encoder.Encode(c)
}

func ReadConfig() (config, error) {
	c := config{}
	// A common use case is to get a private config folder for your app to
	// place its settings files into, that are specific to the local user.
	configPath := configdir.LocalConfig("hlp")
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		return config{}, fmt.Errorf("cannot read path: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, "configuration.json")
	if _, err = os.Stat(configFile); err != nil {
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
	return c, nil
}
