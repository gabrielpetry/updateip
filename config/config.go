package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Log struct {
		Level string
		Path  string
	}
	Provider struct {
		Cloudflare struct {
			APIKey   string
			APIEmail string
			Hostname string
		}
	}
	// Entries []struct {
	// 	Entry     string
	// 	Proxied   bool
	// 	Interface string
	// }
	Ifaces struct {
		Regex struct {
			Name string
			Addr string
		}
	}
	Output string
}

func (c *Config) New() *Config {
	return c
}

var singleInstance *Config

var lock = &sync.Mutex{}
var systemDir = "/etc/go-dhcpdump"

func init() {
	os.MkdirAll(systemDir, os.ModeDir)
}

func (c *Config) parseConfig() *Config {
	// viper config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/updateip/")
	viper.AddConfigPath("$HOME/.config/updateip")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var configuration Config
	err = viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	flag.String("output", "json", "output mode: json or bash")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	configuration.Output = viper.GetString("output") // retrieve value from viper

	return &configuration
}

func GetInstance() *Config {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()

		if singleInstance == nil {
			conf := Config{}
			c := conf.New()
			singleInstance = c.parseConfig()
		}
	}
	return singleInstance
}
