package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Profile 单个账号配置
type Profile struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint,omitempty"`
}

// MultiConfig 多账号配置文件结构
type MultiConfig struct {
	Current  string             `json:"current"`
	Profiles map[string]Profile `json:"profiles"`
}

// CliConfig 运行时配置（已解析的当前账号）
type CliConfig struct {
	SecretID    string `json:"secret_id"`
	SecretKey   string `json:"secret_key"`
	Region      string `json:"region"`
	Endpoint    string `json:"endpoint,omitempty"`
	CurrentAlias string `json:"-"` // 当前账号别名，不序列化
}

// legacyConfig 用于检测旧格式
type legacyConfig struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint,omitempty"`
	// 新格式字段，用于区分
	Current  *string            `json:"current,omitempty"`
	Profiles map[string]Profile `json:"profiles,omitempty"`
}

func GetConfigDir() string {
	if dir := os.Getenv("CLS_CLI_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cls-cli"
	}
	return filepath.Join(home, ".cls-cli")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

// LoadRawConfig 加载完整多账号配置，自动迁移旧格式
func LoadRawConfig() (*MultiConfig, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("未找到配置文件，请先运行 cls-cli config init 或 cls-cli config add")
		}
		return nil, err
	}

	// 先尝试解析为 legacyConfig 判断格式
	var legacy legacyConfig
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("配置文件格式错误: %w", err)
	}

	// 如果有 profiles 字段，是新格式
	if legacy.Profiles != nil {
		var mc MultiConfig
		if err := json.Unmarshal(data, &mc); err != nil {
			return nil, fmt.Errorf("配置文件格式错误: %w", err)
		}
		return &mc, nil
	}

	// 旧格式迁移
	if legacy.SecretID != "" {
		mc := &MultiConfig{
			Current: "default",
			Profiles: map[string]Profile{
				"default": {
					SecretID:  legacy.SecretID,
					SecretKey: legacy.SecretKey,
					Region:    legacy.Region,
					Endpoint:  legacy.Endpoint,
				},
			},
		}
		// 自动保存迁移后的新格式
		if saveErr := SaveRawConfig(mc); saveErr == nil {
			fmt.Fprintf(os.Stderr, "已将旧格式配置自动迁移为多账号格式（别名: default）\n")
		}
		return mc, nil
	}

	return nil, fmt.Errorf("配置文件为空或格式不识别，请运行 cls-cli config add")
}

// SaveRawConfig 保存完整多账号配置
func SaveRawConfig(mc *MultiConfig) error {
	dir := GetConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(mc, "", "  ")
	if err != nil {
		return err
	}
	path := GetConfigPath()
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// LoadConfig 加载当前账号配置（兼容旧调用方式），支持环境变量覆盖
func LoadConfig() (*CliConfig, error) {
	mc, err := LoadRawConfig()
	if err != nil {
		return nil, err
	}

	current := mc.Current
	if current == "" {
		current = "default"
	}

	profile, ok := mc.Profiles[current]
	if !ok {
		return nil, fmt.Errorf("当前账号 [%s] 不存在，请运行 cls-cli config use <别名> 切换账号", current)
	}

	cfg := &CliConfig{
		SecretID:    profile.SecretID,
		SecretKey:   profile.SecretKey,
		Region:      profile.Region,
		Endpoint:    profile.Endpoint,
		CurrentAlias: current,
	}

	// 环境变量优先级最高
	if envID := os.Getenv("TENCENTCLOUD_SECRET_ID"); envID != "" {
		cfg.SecretID = envID
	}
	if envKey := os.Getenv("TENCENTCLOUD_SECRET_KEY"); envKey != "" {
		cfg.SecretKey = envKey
	}
	if envRegion := os.Getenv("CLS_DEFAULT_REGION"); envRegion != "" {
		cfg.Region = envRegion
	}

	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("SecretID 或 SecretKey 未配置，请运行 cls-cli config add")
	}
	return cfg, nil
}

// SaveConfig 保存单个配置（兼容旧调用方式，写入 default profile）
func SaveConfig(cfg *CliConfig) error {
	// 尝试加载现有配置
	mc, err := LoadRawConfig()
	if err != nil {
		// 没有现有配置，创建新的
		mc = &MultiConfig{
			Current:  "default",
			Profiles: map[string]Profile{},
		}
	}

	if mc.Current == "" {
		mc.Current = "default"
	}

	// 写入 default profile
	mc.Profiles["default"] = Profile{
		SecretID:  cfg.SecretID,
		SecretKey: cfg.SecretKey,
		Region:    cfg.Region,
		Endpoint:  cfg.Endpoint,
	}

	return SaveRawConfig(mc)
}
