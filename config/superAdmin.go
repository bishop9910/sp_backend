package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// SuperAdminConfig 超级管理员配置结构
type SuperAdminConfig struct {
	Username string `mapstructure:"username" json:"username" yaml:"username"`
}

// superAdminWrapper 用于解析 YAML 顶层键
type superAdminWrapper struct {
	SuperAdmin SuperAdminConfig `mapstructure:"super_admin" json:"super_admin" yaml:"super_admin"`
}

var (
	_superAdminInstance *SuperAdminConfig
	_superAdminOnce     sync.Once
	_defaultUsername    = "bishop9910"
	_configFileName     = "super_admin"
	_configFileType     = "yaml"
	_configPaths        = []string{
		"./config",
		"../config",
		"/etc/sp_backend",
		"",
	}
)

// GetSuperAdmin 获取超级管理员配置（单例模式 + 懒加载）
// 返回 *SuperAdminConfig 或 error
func GetSuperAdmin() (*SuperAdminConfig, error) {
	var err error
	_superAdminOnce.Do(func() {
		_superAdminInstance, err = loadSuperAdminConfig()
	})
	return _superAdminInstance, err
}

// loadSuperAdminConfig 内部加载逻辑
func loadSuperAdminConfig() (*SuperAdminConfig, error) {
	v := viper.New()
	v.SetConfigName(_configFileName)
	v.SetConfigType(_configFileType)

	// 添加多个配置搜索路径
	for _, path := range _configPaths {
		v.AddConfigPath(path)
	}

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在：使用默认值并尝试创建
			fmt.Printf("[%s.%s] not found, using default username: %s\n",
				_configFileName, _configFileType, _defaultUsername)
			_ = createDefaultConfigFile()
			return &SuperAdminConfig{Username: _defaultUsername}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置到包装结构
	var wrapper superAdminWrapper
	if err := v.Unmarshal(&wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 校验并填充默认值
	username := wrapper.SuperAdmin.Username
	if username == "" {
		username = _defaultUsername
		fmt.Printf("super_admin.username is empty, using default: %s\n", _defaultUsername)
	}

	fmt.Printf("super_admin config loaded: %s\n", username)
	return &SuperAdminConfig{Username: username}, nil
}

// createDefaultConfigFile 创建默认配置文件
func createDefaultConfigFile() error {
	configDir := "./config"
	configPath := filepath.Join(configDir, fmt.Sprintf("%s.%s", _configFileName, _configFileType))

	// 创建配置目录
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 默认 YAML 内容
	defaultYAML := fmt.Sprintf(`# Super Admin Configuration
# Auto-generated - please review and customize

super_admin:
  username: "%s"
  # Additional fields can be added here in the future
`, _defaultUsername)

	if err := os.WriteFile(configPath, []byte(defaultYAML), 0644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	fmt.Printf("created default config: %s\n", configPath)
	return nil
}

// ============ 便捷方法 ============

// IsSuperAdmin 判断指定用户名是否为超级管理员
// 返回 isSuperAdmin (出错默认判断是否为bishop9910)
func IsSuperAdmin(username string) bool {
	cfg, err := GetSuperAdmin()
	if err != nil {
		return username == "bishop9910"
	}
	return cfg.Username == username
}

// GetSuperAdminUsername 直接获取超级管理员用户名字符串
// 加载失败时返回默认值 "bishop9910"
func GetSuperAdminUsername() string {
	cfg, err := GetSuperAdmin()
	if err != nil {
		fmt.Printf("GetSuperAdminUsername fallback to default: %s (error: %v)\n",
			_defaultUsername, err)
		return _defaultUsername
	}
	return cfg.Username
}

func InitSuperAdminConfig() {

	_, err := GetSuperAdmin()
	if err != nil {
		fmt.Printf("[WARN] super_admin config preload failed: %v\n", err)
	}

	fmt.Printf("config/superAdmin module initialized\n")
}
