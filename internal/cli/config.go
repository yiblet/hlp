package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	configpkg "github.com/yiblet/hlp/internal/config"
	"github.com/yiblet/hlp/internal/xio"
)

type ConfigCmd struct {
	Set  *ConfigSetCmd  `arg:"subcommand"`
	Get  *ConfigGetCmd  `arg:"subcommand"`
	Path *ConfigPathCmd `arg:"subcommand"`
}

func (c *ConfigCmd) Execute(ctx context.Context, config *configpkg.Config) error {
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
		if err := enc.Encode(config); err != nil {
			return err
		}

		fmt.Printf("%s", buf.String())
		return nil
	}
}

type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Execute(ctx context.Context, config *configpkg.Config) error {
	fmt.Printf("%s\n", configpkg.Path())
	return nil
}

type ConfigGetCmd struct {
	Model             *struct{} `arg:"subcommand:model"`
	OpenAIAPIKey      *struct{} `arg:"subcommand:openai_api_key"`
	OpenAIAPIEndpoint *struct{} `arg:"subcommand:openai_api_endpoint"`
}

func (c *ConfigGetCmd) Execute(ctx context.Context, config *configpkg.Config) error {
	switch {
	case c.Model != nil:
		return executeGet(config, modelKeyValue{})
	case c.OpenAIAPIKey != nil:
		return executeGet(config, openaiKeyValue{})
	case c.OpenAIAPIEndpoint != nil:
		return executeGet(config, openaiEndpointValue{})
	default:
		return xio.WriteHelp(c, os.Stderr)
	}
}

type ConfigSetCmd struct {
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

func (c *ConfigSetCmd) Execute(ctx context.Context, config *configpkg.Config) error {
	switch {
	case c.Model != nil:
		return executeSet(config, modelKeyValue{}, c.Model.Model)
	case c.OpenAIAPIKey != nil:
		return executeSet(config, openaiKeyValue{}, c.OpenAIAPIKey.OpenAIAPIKey)
	case c.OpenAIAPIEndpoint != nil:
		return executeSet(config, openaiEndpointValue{}, c.OpenAIAPIEndpoint.OpenAIAPIEndpoint)
	default:
		return xio.WriteHelp(c, os.Stderr)
	}
}

type configValue interface {
	set(config *configpkg.Config, value string) error
	get(config *configpkg.Config) string
	name() string
}

type openaiEndpointValue struct{}

func (openaiEndpointValue) set(config *configpkg.Config, value string) error {
	config.OpenAIAPIEndpoint = value
	return nil
}

func (openaiEndpointValue) get(config *configpkg.Config) string {
	return config.OpenAIAPIEndpoint
}

func (openaiEndpointValue) fromEnv() string {
	val, _ := os.LookupEnv("OPENAI_API_ENDPOINT")
	return val
}

func (openaiEndpointValue) name() string {
	return "openai api endpoint"
}

type openaiKeyValue struct{}

func (openaiKeyValue) set(config *configpkg.Config, value string) error {
	config.OpenAIAPIKey = value
	return nil
}

func (openaiKeyValue) get(config *configpkg.Config) string {
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

func (modelKeyValue) set(config *configpkg.Config, value string) error {
	config.DefaultModel = value
	return nil
}

func (modelKeyValue) get(config *configpkg.Config) string {
	return config.DefaultModel
}

func (modelKeyValue) name() string {
	return "openai model"
}

func executeSet(config *configpkg.Config, configVal configValue, value string) error {
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

func executeGet(config *configpkg.Config, configVal configValue) error {
	if value := configVal.get(config); value != "" {
		fmt.Printf("%s\n", value)
		return nil
	}

	return fmt.Errorf("value is empty")
}
