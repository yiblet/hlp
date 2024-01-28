package main

import (
	"context"
	"fmt"
	"strings"
)

type authCmd struct {
	OpenAIAPIKey string `arg:"env:OPENAI_API_KEY"`
}

func (c *authCmd) Execute(ctx context.Context) error {
	if c.OpenAIAPIKey == "" {
		fmt.Printf("The openai api key was not passed in via env or command line argument. Enter it in the following line:\n")
		var key string
		_, err := fmt.Scanf("%s\n", &key)
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("invalid openai key")
		}
		c.OpenAIAPIKey = key
	}

	cmd := authConfig{
		OpenAIAPIKey: c.OpenAIAPIKey,
	}

	fmt.Printf("saving openai api key...\n")
	err := cmd.Write()
	if err != nil {
		return err
	}

	fmt.Printf("api key stored in cofig\n")
	return nil
}
