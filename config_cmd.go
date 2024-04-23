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
	Set  *configSetCmd  `arg:"subcommand"`
	Get  *configGetCmd  `arg:"subcommand"`
	Path *configPathCmd `arg:"subcommand"`
}

func (c *configCmd) Execute(ctx context.Context, config *config) error {
	switch {
	case c.Set != nil:
		return c.Set.Execute(ctx, config)
	case c.Get != nil:
		return c.Get.Execute(ctx, config)
	case c.Path != nil:
		return c.Path.Execute(ctx, config)
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

type configPathCmd struct{}

func (c *configPathCmd) Execute(ctx context.Context, config *config) error {
	fmt.Printf("%s\n", getConfigPath())
	return nil
}

type configGetCmd struct {
	Model *struct {
	} `arg:"subcommand:model"`
	OpenAIAPIKey *struct {
	} `arg:"subcommand:openai_api_key"`
	OpenAIAPIEndpoint *struct {
	} `arg:"subcommand:openai_api_endpoint"`
}

func (c *configGetCmd) Execute(ctx context.Context, config *config) error {
	switch {
	case c.Model != nil:
		return executeGet(config, modelKeyValue{})
	case c.OpenAIAPIKey != nil:
		return executeGet(config, openaiKeyValue{})
	case c.OpenAIAPIEndpoint != nil:
		return executeGet(config, openaiEndpointValue{})
	default:
		return writeHelp(c, os.Stderr)
	}
}

type configSetCmd struct {
	Model *struct {
		Model string `arg:"positional"`
	} `arg:"subcommand:model"`
	OpenAIAPIKey *struct {
		OpenAIAPIKey string `arg:"positional"`
	} `arg:"subcommand:openai_api_key"`
	OpenAIAPIEndpoint *struct {
		OpenAIAPIEndpoint string `arg:"positional"`
	} `arg:"subcommand:openai_api_endpoint"`
}

func (c *configSetCmd) Execute(ctx context.Context, config *config) error {
	switch {
	case c.Model != nil:
		return executeSet(config, modelKeyValue{}, c.Model.Model)
	case c.OpenAIAPIKey != nil:
		return executeSet(config, openaiKeyValue{}, c.OpenAIAPIKey.OpenAIAPIKey)
	case c.OpenAIAPIEndpoint != nil:
		return executeSet(config, openaiEndpointValue{}, c.OpenAIAPIEndpoint.OpenAIAPIEndpoint)
	default:
		return writeHelp(c, os.Stderr)
	}
}

type configValue interface {
	set(config *config, value string) error
	get(config *config) string
	name() string
}

type openaiEndpointValue struct{}

func (openaiEndpointValue) set(config *config, value string) error {
	config.OpenaiAPIEndpoint = value
	return nil
}

func (openaiEndpointValue) get(config *config) string {
	return config.OpenaiAPIEndpoint
}

func (openaiEndpointValue) fromEnv() string {
	val, _ := os.LookupEnv("OPENAI_API_ENDPOINT")
	return val
}

func (openaiEndpointValue) name() string {
	return "openai api endpoint"
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

func executeSet(config *config, configVal configValue, value string) error {
	if env, ok := configVal.(interface{ fromEnv() string }); value == "" && ok {
		value = strings.TrimSpace(env.fromEnv())
		if value != "" {
			fmt.Printf("%s is set from env\n", configVal.name())
		}
	}

	if value == "" {
		_, hasEnv := configVal.(interface{ fromEnv() string })
		if hasEnv {
			fmt.Printf("The %s was not passed in via env or command line argument. Enter it in the following line:\n", configVal.name())
		} else {
			fmt.Printf("The %s was not passed in via command line argument. Enter it in the following line:\n", configVal.name())
		}

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
	if err := config.Write(); err != nil {
		return err
	}

	fmt.Printf("%s stored in cofig\n", configVal.name())
	return nil
}

func executeGet(config *config, configVal configValue) error {
	if value := configVal.get(config); value != "" {
		fmt.Printf("%s\n", value)
		return nil
	}

	return fmt.Errorf("value is empty")
}
