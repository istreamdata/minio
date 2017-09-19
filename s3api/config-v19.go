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
	"sync"

	"github.com/minio/minio/pkg/quick"
)

var (
	// serverConfig server config.
	serverConfig *serverConfigV19
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
