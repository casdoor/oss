// Copyright 2023 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synology_test

import (
	"testing"

	"github.com/casdoor/oss/synology"
	"github.com/casdoor/oss/tests"
	"github.com/jinzhu/configor"
)

type Config struct {
	AccessID  string
	AccessKey string
	Region    string
	Bucket    string
	Endpoint  string
}

type AppConfig struct {
	Private Config
	Public  Config
}

var client *synology.Client
var privateClient *synology.Client

func init() {
	config := AppConfig{}
	configor.New(&configor.Config{ENVPrefix: "SYNOLOGY"}).Load(&config)
	if len(config.Private.AccessID) == 0 {
		return
	}

	client = synology.New(&synology.Config{
		AccessID:  config.Public.AccessID,
		AccessKey: config.Public.AccessKey,
		Endpoint:  config.Public.Endpoint,

	})
	privateClient = synology.New(&synology.Config{
		AccessID:   config.Private.AccessID,
		AccessKey:  config.Private.AccessKey,
		Endpoint:   config.Private.Endpoint,
	})
}

func TestAll(t *testing.T) {
	if client == nil {
		t.Skip(`skip because of no config:


			`)
	}
	clis := []*synology.Client{client, privateClient}
	for _, cli := range clis {
		tests.TestAll(cli, t)
	}
}
