/*
 * Minio Cloud Storage, (C) 2016, 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package s3api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/minio/minio/pkg/quick"
	"github.com/tidwall/gjson"
)

// Config version
const v19 = "19"

var (
	// serverConfig server config.
	serverConfig   *serverConfigV19
	serverConfigMu sync.RWMutex
)

// serverConfigV19 server configuration version '19' which is like
// version '18' except it adds support for MQTT notifications.
type serverConfigV19 struct {
	sync.RWMutex
	Version string `json:"version"`

	// S3 API configuration.
	Credential credential `json:"credential"`
	Region     string     `json:"region"`

	// Additional error logging configuration.
	Logger *loggers `json:"logger"`
}

// GetVersion get current config version.
func (s *serverConfigV19) GetVersion() string {
	s.RLock()
	defer s.RUnlock()

	return s.Version
}

// SetRegion set a new region.
func (s *serverConfigV19) SetRegion(region string) {
	s.Lock()
	defer s.Unlock()

	// Save new region.
	s.Region = region
}

// GetRegion get current region.
func (s *serverConfigV19) GetRegion() string {
	s.RLock()
	defer s.RUnlock()

	return s.Region
}

// SetCredentials set new credentials. SetCredential returns the previous credential.
func (s *serverConfigV19) SetCredential(creds credential) (prevCred credential) {
	s.Lock()
	defer s.Unlock()

	// Save previous credential.
	prevCred = s.Credential

	// Set updated credential.
	s.Credential = creds

	// Return previous credential.
	return prevCred
}

// GetCredentials get current credentials.
func (s *serverConfigV19) GetCredential() credential {
	s.RLock()
	defer s.RUnlock()

	return s.Credential
}

// Save config.
func (s *serverConfigV19) Save() error {
	s.RLock()
	defer s.RUnlock()

	// Save config file.
	return quick.Save(getConfigFile(), s)
}

func newServerConfigV19() *serverConfigV19 {
	srvCfg := &serverConfigV19{
		Version:    v19,
		Credential: mustGetNewCredential(),
		Region:     globalMinioDefaultRegion,
		Logger:     &loggers{},
	}

	// Enable console logger by default on a fresh run.
	srvCfg.Logger.Console = NewConsoleLogger()

	return srvCfg
}

// newConfig - initialize a new server config, saves env parameters if
// found, otherwise use default parameters
func newConfig() error {
	// Initialize server config.
	srvCfg := newServerConfigV19()

	// If env is set override the credentials from config file.
	if globalIsEnvCreds {
		srvCfg.SetCredential(globalActiveCred)
	}

	if globalIsEnvRegion {
		srvCfg.SetRegion(globalServerRegion)
	}

	// hold the mutex lock before a new config is assigned.
	// Save the new config globally.
	// unlock the mutex.
	serverConfigMu.Lock()
	serverConfig = srvCfg
	serverConfigMu.Unlock()

	// Save config into file.
	return serverConfig.Save()
}

// doCheckDupJSONKeys recursively detects duplicate json keys
func doCheckDupJSONKeys(key, value gjson.Result) error {
	// Key occurrences map of the current scope to count
	// if there is any duplicated json key.
	keysOcc := make(map[string]int)

	// Holds the found error
	var checkErr error

	// Iterate over keys in the current json scope
	value.ForEach(func(k, v gjson.Result) bool {
		// If current key is not null, check if its
		// value contains some duplicated keys.
		if k.Type != gjson.Null {
			keysOcc[k.String()]++
			checkErr = doCheckDupJSONKeys(k, v)
		}
		return checkErr == nil
	})

	// Check found err
	if checkErr != nil {
		return errors.New(key.String() + " => " + checkErr.Error())
	}

	// Check for duplicated keys
	for k, v := range keysOcc {
		if v > 1 {
			return errors.New(key.String() + " => `" + k + "` entry is duplicated")
		}
	}

	return nil
}

// Check recursively if a key is duplicated in the same json scope
// e.g.:
//  `{ "key" : { "key" ..` is accepted
//  `{ "key" : { "subkey" : "val1", "subkey": "val2" ..` throws subkey duplicated error
func checkDupJSONKeys(json string) error {
	// Parse config with gjson library
	config := gjson.Parse(json)

	// Create a fake rootKey since root json doesn't seem to have representation
	// in gjson library.
	rootKey := gjson.Result{Type: gjson.String, Str: minioConfigFile}

	// Check if loaded json contains any duplicated keys
	return doCheckDupJSONKeys(rootKey, config)
}

// getValidConfig - returns valid server configuration
func getValidConfig() (*serverConfigV19, error) {
	srvCfg := &serverConfigV19{
		Region: globalMinioDefaultRegion,
	}

	configFile := getConfigFile()
	if _, err := quick.Load(configFile, srvCfg); err != nil {
		return nil, err
	}

	if srvCfg.Version != v19 {
		return nil, fmt.Errorf("configuration version mismatch. Expected: ‘%s’, Got: ‘%s’", v19, srvCfg.Version)
	}

	// Load config file json and check for duplication json keys
	jsonBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	if err = checkDupJSONKeys(string(jsonBytes)); err != nil {
		return nil, err
	}

	// Validate credential fields only when
	// they are not set via the environment

	// Error out if global is env credential is not set and config has invalid credential
	if !globalIsEnvCreds && !srvCfg.Credential.IsValid() {
		return nil, errors.New("invalid credential in config file " + configFile)
	}

	// Validate logger field
	if err = srvCfg.Logger.Validate(); err != nil {
		return nil, err
	}

	return srvCfg, nil
}

// loadConfig - loads a new config from disk, overrides params from env
// if found and valid
func loadConfig() error {
	srvCfg, err := getValidConfig()
	if err != nil {
		return err
	}

	// If env is set override the credentials from config file.
	if globalIsEnvCreds {
		srvCfg.SetCredential(globalActiveCred)
	}

	if globalIsEnvRegion {
		srvCfg.SetRegion(globalServerRegion)
	}

	// hold the mutex lock before a new config is assigned.
	serverConfigMu.Lock()
	serverConfig = srvCfg
	if !globalIsEnvCreds {
		globalActiveCred = serverConfig.GetCredential()
	}
	if !globalIsEnvRegion {
		globalServerRegion = serverConfig.GetRegion()
	}
	serverConfigMu.Unlock()

	return nil
}
