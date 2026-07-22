package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bind        string `yaml:"bind"`
	Port        int    `yaml:"port"`
	Databases   int    `yaml:"databases"`
	AofFile     string `yaml:"aof_file"`
	LogFile     string `yaml:"log_file"`
	LogLevel    string `yaml:"log_level"`
	RequirePass string `yaml:"requirepass"`
	WebAdmin    bool   `yaml:"web_admin"`
	WebBind     string `yaml:"web_bind"`
	WebPort     int    `yaml:"web_port"`
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
		WebAdmin:  true,
		WebBind:   "0.0.0.0",
		WebPort:   6390,
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

	tmpl := `# Godis configuration file
# Automatically generated. Modify as needed.

# 监听地址，0.0.0.0 表示监听所有网络接口
# 默认值: {{.Bind}}
bind: {{.Bind}}

# 监听端口
# 默认: {{.Port}}
port: {{.Port}}

# 逻辑数据库数量，每个数据库独立隔离
# 默认: {{.Databases}}
databases: {{.Databases}}

# AOF 持久化文件路径
# 默认: {{.AofFile}}
aof_file: {{.AofFile}}

# 日志文件路径，支持 lumberjack 滚动切割
# 默认: {{.LogFile}}
log_file: {{.LogFile}}

# 日志级别：debug / info / warn / error
# 默认: {{.LogLevel}}
log_level: {{.LogLevel}}

# Web 管理后台配置
# 是否启用 Web 管理后台（默认 true）
web_admin: {{.WebAdmin}}
# Web 监听地址（默认 0.0.0.0）
web_bind: {{.WebBind}}
# Web 监听端口（默认 6390）
web_port: {{.WebPort}}

# 认证密码，留空表示不启用认证
# 示例: requirepass: mypassword
# {{.RequirePass}}
`
	var buf bytes.Buffer
	template.Must(template.New("config").Parse(tmpl)).Execute(&buf, defaults())
	_, err = file.WriteString(buf.String())
	return err
}

// load 解析 YAML 配置文件
func load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &Global)
}
