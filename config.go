package nopaste

import (
	"errors"
	"io/ioutil"
	"log"
	"path"

	goconfig "github.com/kayac/go-config"
	"gopkg.in/yaml.v2"
)

type Config struct {
	BaseURL  string       `yaml:"base_url"`
	Listen   string       `yaml:"listen"`
	DataDir  string       `yaml:"data_dir"`
	S3       *S3Config    `yaml:"s3"`
	IRC      *IRCConfig   `yaml:"irc"`
	Slack    *SlackConfig `yaml:"slack"`
	filePath string
	Channels []string `yaml:"channels"`
	storages []Storage
}

type S3Config struct {
	Bucket    string `yaml:"bucket"`
	KeyPrefix string `yaml:"key_prefix"`
}

type IRCConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Secure   bool   `yaml:"secure"`
	Debug    bool   `yaml:"debug"`
	Password string `yaml:"password"`
	Nick     string `yaml:"nick"`
}

type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

func (c *Config) AddChannel(channel string) {
	for _, _channel := range c.Channels {
		if channel == _channel {
			return
		}
	}
	c.Channels = append(c.Channels, channel)
	err := c.Save()
	if err != nil {
		log.Println("[warn]", err)
	}
	return
}

func (c *Config) Save() error {
	log.Println("[info] saving config file", c.filePath)
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.filePath, data, 0644)
}

func (c *Config) SetFilePath(path string) {
	c.filePath = path
}

func (c *Config) Storages() []Storage {
	return c.storages
}

func LoadConfig(file string) (*Config, error) {
	log.Println("[info] loading config file", file)
	c := Config{filePath: file}
	err := goconfig.LoadWithEnv(&c, file)
	if err != nil {
		return nil, err
	}

	if c.S3 != nil {
		log.Printf("[info] using S3 storage s3://%s", path.Join(c.S3.Bucket, c.S3.KeyPrefix))
		c.storages = append(c.storages, NewS3Storage(c.S3))
	}
	if c.DataDir != "" {
		log.Printf("[info] using local storage %s", c.DataDir)
		c.storages = append(c.storages, NewLocalStorage(c.DataDir))
	}
	if len(c.storages) == 0 {
		return nil, errors.New("s3 or data_dir is required")
	}

	return &c, nil
}
