package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bind      string `yaml:"bind"`
	Port      int    `yaml:"port"`
	Databases int    `yaml:"databases"`
	AofFile   string `yaml:"aof_file"`
	LogFile   string `yaml:"log_file"`
	LogLevel  string `yaml:"log_level"`
}

var Global Config
var configPath string

func defaults() Config {
	return Config{
		Bind:      "0.0.0.0",
		Port:      6379,
		Databases: 16,
		AofFile:   "./data/godis.aof",
		LogFile:   "./logs/godis.log",
		LogLevel:  "info",
	}
}

// Init 加载配置文件，不存在则自动生成默认配置
func Init(filename string) error {
	configPath = filename
	Global = defaults()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := generate(filename); err != nil {
			return fmt.Errorf("failed to generate config file: %v", err)
		}
		return nil
	}

	return load(filename)
}

// Save 将当前配置写回配置文件
func Save() error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	header := "# Godis configuration file\n\n"
	if _, err := file.WriteString(header); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(Global)
}

// generate 生成默认配置文件
func generate(filename string) error {
	dir := filepath.Dir(filename)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	d := defaults()
	header := "# Godis configuration file\n# Automatically generated. Modify as needed.\n\n"
	if _, err := file.WriteString(header); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(d)
}

// load 解析 YAML 配置文件
func load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &Global)
}
