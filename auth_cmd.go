package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type configCmd struct {
	Set *configSetCmd `arg:"subcommand"`
}

func (c *configCmd) Execute(ctx context.Context, config *config) error {
	switch {
	case c.Set != nil:
		return c.Set.Execute(ctx, config)
	default:
		buf := bytes.NewBuffer([]byte{})
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		err := enc.Encode(config)
		if err != nil {
			return err
		}

		fmt.Printf("%s", buf.String())
		return nil
	}
}

type configSetCmd struct {
	Model *struct {
		Model string `arg:"positional"`
	} `arg:"subcommand"`
	OpenAIAPIKey *struct {
		OpenAIAPIKey string `arg:"positional"`
	} `arg:"subcommand"`
}

func (c *configSetCmd) Execute(ctx context.Context, config *config) error {
	switch {
	case c.Model != nil:
		return executeSet(ctx, config, modelKeyValue{}, c.Model.Model)
	case c.OpenAIAPIKey != nil:
		return executeSet(ctx, config, openaiKeyValue{}, c.OpenAIAPIKey.OpenAIAPIKey)
	default:
		return WriteHelp(c, os.Stderr)
	}
}

type configGenericSet[val configValue] struct {
	Value string `arg:"positional"`
}

type configValue interface {
	set(config *config, value string) error
	get(config *config) string
	name() string
}

type openaiKeyValue struct{}

func (openaiKeyValue) set(config *config, value string) error {
	config.OpenAIAPIKey = value
	return nil
}

func (openaiKeyValue) get(config *config) string {
	return config.OpenAIAPIKey
}

func (openaiKeyValue) fromEnv() string {
	val, _ := os.LookupEnv("OPENAI_API_KEY")
	return val
}

func (openaiKeyValue) name() string {
	return "openai api key"
}

type modelKeyValue struct{}

func (modelKeyValue) set(config *config, value string) error {
	config.DefaultModel = value
	return nil
}

func (modelKeyValue) get(config *config) string {
	return config.DefaultModel
}

func (modelKeyValue) name() string {
	return "openai model"
}

func executeSet(ctx context.Context, config *config, configVal configValue, value string) error {
	if env, ok := configVal.(interface{ fromEnv() string }); value == "" && ok {
		value = strings.TrimSpace(env.fromEnv())
		if value != "" {
			fmt.Printf("%s is set from env", configVal.name())
		}
	}

	if value == "" {
		fmt.Printf("The %s was not passed in via env or command line argument. Enter it in the following line:\n", configVal.name())
		var key string
		_, err := fmt.Scanf("%s\n", &key)
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("invalid %s", configVal.name())
		}
		value = key
	}

	if err := configVal.set(config, value); err != nil {
		return err
	}

	fmt.Printf("saving %s...\n", configVal.name())
	if err := WriteConfig(config); err != nil {
		return err
	}

	fmt.Printf("%s stored in cofig\n", configVal.name())
	return nil
}
