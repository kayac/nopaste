package nopaste

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v1"
)

type Config struct {
	BaseURL  string    `yaml:"base_url"`
	Listen   string    `yaml:"listen"`
	DataDir  string    `yaml:"data_dir"`
	IRC      IRCConfig `yaml:"irc"`
	filePath string
}

type IRCConfig struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Secure   bool     `yaml:"secure"`
	Debug    bool     `yaml:"debug"`
	Password string   `yaml:"password"`
	Nick     string   `yaml:"nick"`
	Channels []string `yaml:"channels"`
}

func (c *Config) AddChannel(channel string) {
	for _, _channel := range c.IRC.Channels {
		if channel == _channel {
			return
		}
	}
	c.IRC.Channels = append(c.IRC.Channels, channel)
	err := c.Save()
	if err != nil {
		log.Println(err)
	}
	return
}

func (c *Config) Save() error {
	log.Println("saving config file", c.filePath)
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.filePath, data, 0644)
}

func (c *Config) SetFilePath(path string) {
	c.filePath = path
}

func LoadConfig(file string) (*Config, error) {
	log.Println("loading config file", file)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	c := Config{filePath: file}
	err = yaml.Unmarshal(content, &c)
	return &c, nil
}
