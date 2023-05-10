package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type LogLevel log.Level

const (
	DebugLevel = (LogLevel)(log.DebugLevel)
	InfoLevel  = (LogLevel)(log.InfoLevel)
	WarnLevel  = (LogLevel)(log.WarnLevel)
	ErrorLevel = (LogLevel)(log.ErrorLevel)
)

func (ll *LogLevel) Set(value string) error {
	level, err := log.ParseLevel(value)
	*ll = (LogLevel)(level)
	return err
}

func (ll *LogLevel) String() string {
	return (log.Level)(*ll).String()
}

func (ll *LogLevel) Type() string {
	return "LogLevel"
}

func (ll *LogLevel) Level() log.Level {
	return (log.Level)(*ll)
}

func LogLevelDecodeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(InfoLevel) {
			return data, nil
		}

		level, err := log.ParseLevel(data.(string))
		if err != nil {
			return InfoLevel, err
		}
		return (LogLevel)(level), nil
	}
}

type CloudbuildOpts struct {
	// General options:
	ConfigPath      string   `mapstructure:"config-path"`
	LogLevel        LogLevel `mapstructure:"log-level"`
	HTTPBindAddress net.IP   `mapstructure:"listen-ip"`
	HTTPBindPort    uint16   `mapstructure:"port"`

	// Database options:
	DatabaseDSN  string `mapstructure:"database-dsn"`
	DatabaseHost string `mapstructure:"database-host"`

	// Build options:
	BuildImage       string `mapstructure:"build-img"`
	SourceRepository string `mapstructure:"src-repo"`

	// Storage options:
	DownloadURL            string `mapstructure:"download-url"`
	StorageType            string `mapstructure:"storage-type"`
	StoragePath            string `mapstructure:"storage-path"`
	StorageS3Bucket        string `mapstructure:"s3-bucket"`
	StorageS3URL           string `mapstructure:"s3-url"`
	StorageS3HostImmutable bool   `mapstructure:"s3-url-immutable"`
	StorageS3AccessKey     string `mapstructure:"s3-access-key"`
	StorageS3SecretKey     string `mapstructure:"s3-secret-key"`

	Viper *viper.Viper
}

func NewOpts(v *viper.Viper) *CloudbuildOpts {
	return &CloudbuildOpts{
		Viper:            v,
		LogLevel:         InfoLevel,
		HTTPBindPort:     3000,
		BuildImage:       "ghcr.io/edgetx/edgetx-builder",
		SourceRepository: "https://github.com/EdgeTX/edgetx.git",
		DownloadURL:      "http://localhost:3000",
		StorageType:      "FILE_SYSTEM_STORAGE",
		StoragePath:      "/tmp",
	}
}

func (o *CloudbuildOpts) BindCliOpts(c *cobra.Command) {
	c.PersistentFlags().Var(&o.LogLevel, "log-level", "Log level")
	c.PersistentFlags().StringVar(&o.ConfigPath, "config-path", o.ConfigPath, "Config path")
}

func (o *CloudbuildOpts) BindDBOpts(c *cobra.Command) {
	c.PersistentFlags().StringVar(
		&o.DatabaseDSN, "database-dsn", o.DatabaseDSN, "Database DSN")
	c.PersistentFlags().StringVar(
		&o.DatabaseHost, "database-host", o.DatabaseHost, "Database host")
}

func (o *CloudbuildOpts) BindStorageOpts(c *cobra.Command) {
	c.PersistentFlags().StringVar(
		&o.StorageType, "storage-type", o.StorageType,
		"Storage type (S3 or FILE_SYSTEM_STORAGE)",
	)
	c.PersistentFlags().StringVar(
		&o.StoragePath, "storage-path", o.StoragePath,
		"Storage path (FILE_SYSTEM_STORAGE)",
	)
	c.PersistentFlags().StringVar(
		&o.StorageS3Bucket, "s3-bucket", o.StorageS3Bucket,
		"S3 bucket",
	)
	c.PersistentFlags().StringVar(
		&o.StorageS3URL, "s3-url", o.StorageS3URL,
		"S3 API URL",
	)
	c.PersistentFlags().BoolVar(
		&o.StorageS3HostImmutable, "s3-url-immutable", o.StorageS3HostImmutable,
		"S3 API URL does not change",
	)
	c.PersistentFlags().StringVar(
		&o.StorageS3AccessKey, "s3-access-key", o.StorageS3AccessKey,
		"Storage access key",
	)
	c.PersistentFlags().StringVar(
		&o.StorageS3SecretKey, "s3-secret-key", o.StorageS3SecretKey,
		"Storage secret key",
	)
}

func (o *CloudbuildOpts) BindAPIOpts(c *cobra.Command) {
	c.Flags().Uint16VarP(&o.HTTPBindPort, "port", "p", o.HTTPBindPort, "HTTP listen port")
	c.Flags().IPVarP(&o.HTTPBindAddress, "listen-ip", "l", net.IPv4zero, "HTTP listen IP")
	c.Flags().StringVarP(
		&o.DownloadURL, "download-url", "u", o.DownloadURL, "Artifact download URL")
}

func (o *CloudbuildOpts) BindWorkerOpts(c *cobra.Command) {
	c.Flags().StringVar(&o.BuildImage, "build-img", o.BuildImage, "Build docker image")
	c.Flags().StringVar(
		&o.SourceRepository, "src-repo", o.SourceRepository, "Source repository")
}

func (o *CloudbuildOpts) Unmarshal() error {
	// Unmarshal config into struct
	return o.Viper.Unmarshal(
		o,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToIPHookFunc(),
				LogLevelDecodeHookFunc(),
			),
		),
	)
}

func (o *CloudbuildOpts) JSON() string {
	bytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func InitConfig(v *viper.Viper) func() {
	return func() {
		v.SetEnvPrefix("ebuild")
		v.AutomaticEnv()

		// This normalizes "-" to an underscore in env names.
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

		configPath := v.GetString("config-path")
		if configPath == "" {
			// Default to looking in the working directory of the running process.
			configPath = "."
		}

		switch strings.ToLower(path.Ext(configPath)) {
		case ".json", ".toml", ".yaml", ".yml":
			v.SetConfigFile(configPath)
		default:
			v.AddConfigPath(configPath)
		}

		if err := v.ReadInConfig(); err != nil && !os.IsNotExist(err) {
			var verr viper.ConfigFileNotFoundError
			if !errors.As(err, &verr) {
				fmt.Println("Can't read config:", err)
				os.Exit(1)
			}
		}

		configFile := v.ConfigFileUsed()
		if configFile != "" {
			log.Debugln("Config filename:", configFile)
		}
	}
}
