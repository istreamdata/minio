package main

import (
	"os"
	"path/filepath"

	"github.com/minio/minio/pkg/quick"
)

/////////////////// Config V1 ///////////////////
type configV1 struct {
	Version         string `json:"version"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

// loadConfigV1 load config
func loadConfigV1() (*configV1, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(configPath, "fsUsers.json")
	if _, err = os.Stat(configFile); err != nil {
		return nil, err
	}
	a := &configV1{}
	a.Version = "1"
	qc, err := quick.New(a)
	if err != nil {
		return nil, err
	}
	if err := qc.Load(configFile); err != nil {
		return nil, err
	}
	return qc.Data().(*configV1), nil
}

/////////////////// Config V2 ///////////////////
type configV2 struct {
	Version     string `json:"version"`
	Credentials struct {
		AccessKeyID     string `json:"accessKeyId"`
		SecretAccessKey string `json:"secretAccessKey"`
		Region          string `json:"region"`
	} `json:"credentials"`
	MongoLogger struct {
		Addr       string `json:"addr"`
		DB         string `json:"db"`
		Collection string `json:"collection"`
	} `json:"mongoLogger"`
	SyslogLogger struct {
		Network string `json:"network"`
		Addr    string `json:"addr"`
	} `json:"syslogLogger"`
	FileLogger struct {
		Filename string `json:"filename"`
	} `json:"fileLogger"`
}

// loadConfigV2 load config version '2'.
func loadConfigV2() (*configV2, error) {
	configFile, err := getConfigFile()
	if err != nil {
		return nil, err
	}
	if _, err = os.Stat(configFile); err != nil {
		return nil, err
	}
	a := &configV2{}
	a.Version = "2"
	qc, err := quick.New(a)
	if err != nil {
		return nil, err
	}
	if err := qc.Load(configFile); err != nil {
		return nil, err
	}
	return qc.Data().(*configV2), nil
}

/////////////////// Config V3 ///////////////////

// backendV3 type.
type backendV3 struct {
	Type  string   `json:"type"`
	Disk  string   `json:"disk,omitempty"`
	Disks []string `json:"disks,omitempty"`
}

// loggerV3 type.
type loggerV3 struct {
	Console struct {
		Enable bool   `json:"enable"`
		Level  string `json:"level"`
	}
	File struct {
		Enable   bool   `json:"enable"`
		Filename string `json:"fileName"`
		Level    string `json:"level"`
	}
	Syslog struct {
		Enable bool   `json:"enable"`
		Addr   string `json:"address"`
		Level  string `json:"level"`
	} `json:"syslog"`
	// Add new loggers here.
}

// configV3 server configuration version '3'.
type configV3 struct {
	Version string `json:"version"`

	// Backend configuration.
	Backend backendV3 `json:"backend"`

	// http Server configuration.
	Addr string `json:"address"`

	// S3 API configuration.
	Credential credential `json:"credential"`
	Region     string     `json:"region"`

	// Additional error logging configuration.
	Logger loggerV3 `json:"logger"`
}

// loadConfigV3 load config version '3'.
func loadConfigV3() (*configV3, error) {
	configFile, err := getConfigFile()
	if err != nil {
		return nil, err
	}
	if _, err = os.Stat(configFile); err != nil {
		return nil, err
	}
	a := &configV3{}
	a.Version = "3"
	qc, err := quick.New(a)
	if err != nil {
		return nil, err
	}
	if err := qc.Load(configFile); err != nil {
		return nil, err
	}
	return qc.Data().(*configV3), nil
}
