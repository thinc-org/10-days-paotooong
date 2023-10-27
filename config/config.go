package config

import "github.com/spf13/viper"

type GrpcConfig struct {
	Port               int    `mapstructure:"port"`
	DbConnectionString string `mapstructure:"connection"`
	Secret             string `mapstructure:"secret"`
}

type ProxyConfig struct {
	Port    int    `mapstruct:"port"`
	GrpcUrl string `mapstructure:"grpc_url"`
}

func LoadGrpcConfig() (*GrpcConfig, error) {
	viper.AddConfigPath("config")
	viper.SetConfigName("grpc.config")
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config GrpcConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadProxyConfig() (*ProxyConfig, error) {
	viper.AddConfigPath("config")
	viper.SetConfigName("proxy.config")
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config ProxyConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
