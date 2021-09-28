package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	WebServer struct {
		Port      string `yaml:"port"`
		Host      string `yaml:"host"`
		TCPClient struct {
			Port          string `yaml:"port"`
			Host          string `yaml:"host"`
			NumConnection int    `yaml:"num_conection"`
		} `yaml:"tcp_client"`
	} `yaml:"web_server"`

	CoreServer struct {
		Port     string `yaml:"port"`
		Host     string `yaml:"host"`
		Database struct {
			DbDriver string `yaml:"dbdriver"`
			DbUser   string `yaml:"dbuser"`
			DbPass   string `yaml:"dbpass"`
			DbName   string `yaml:"dbname"`
		} `yaml:"database"`
		Redis struct {
			Port       string `yaml:"port"`
			Host       string `yaml:"host"`
			Index      int    `yaml:"index"`
			Pass       string `yaml:"pass"`
			ExpireTime int    `yamk:"expire_time"`
		} `yaml:"redis"`
	} `yaml:"core_server"`
}

func LoadConfig() (*Config, error) {
	f, err := os.Open("./config.yml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
