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

package qiniu_test

import (
	"testing"

	"github.com/casdoor/oss/qiniu"
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

var client *qiniu.Client
var privateClient *qiniu.Client

func init() {
	config := AppConfig{}
	configor.New(&configor.Config{ENVPrefix: "QINIU"}).Load(&config)
	if len(config.Private.AccessID) == 0 {
		return
	}

	var err error
	client, err = qiniu.New(&qiniu.Config{
		AccessID:  config.Public.AccessID,
		AccessKey: config.Public.AccessKey,
		Region:    config.Public.Region,
		Bucket:    config.Public.Bucket,
		Endpoint:  config.Public.Endpoint,
	})
	if err != nil {
		panic(err)
	}

	privateClient, err = qiniu.New(&qiniu.Config{
		AccessID:   config.Private.AccessID,
		AccessKey:  config.Private.AccessKey,
		Region:     config.Private.Region,
		Bucket:     config.Private.Bucket,
		Endpoint:   config.Private.Endpoint,
		PrivateURL: true,
	})
	if err != nil {
		panic(err)
	}
}

func TestAll(t *testing.T) {
	if client == nil {
		t.Skip(`skip because of no config:


			`)
	}
	clis := []*qiniu.Client{client, privateClient}
	for _, cli := range clis {
		tests.TestAll(cli, t)
	}
}
