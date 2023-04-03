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

type authConfig struct {
	OpenAIAPIKey string `json:"openai_api_key"`
}

func (c *authConfig) NewClient() (gpt3.Client, error) {
	if err := c.Read(); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: 0,
	}
	client := gpt3.NewClient(c.OpenAIAPIKey, gpt3.WithHTTPClient(httpClient))
	return client, nil
}

func (c *authConfig) Read() error {
	// A common use case is to get a private config folder for your app to
	// place its settings files into, that are specific to the local user.
	configPath := configdir.LocalConfig("hlp")
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		return fmt.Errorf("cannot read path: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, "configuration.json")
	if _, err = os.Stat(configFile); err != nil {
		return err
	}

	// Load the existing file.
	fh, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	decoder := json.NewDecoder(fh)
	return decoder.Decode(c)
}

func (c *authConfig) Write() error {
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
