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

package aliyun_test

import (
	"testing"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/casdoor/oss/aliyun"
	"github.com/casdoor/oss/tests"
	"github.com/jinzhu/configor"
)

type Config struct {
	AccessID  string
	AccessKey string
	Bucket    string
	Endpoint  string
}

type AppConfig struct {
	Private Config
	Public  Config
}

var client, privateClient *aliyun.Client

func init() {
	config := AppConfig{}
	err := configor.New(&configor.Config{ENVPrefix: "ALIYUN"}).Load(&config)
	if err != nil {
		panic(err)
	}

	if config.Private.AccessID == "" {
		panic("No aliyun configuration")
	}

	client = aliyun.New(&aliyun.Config{
		AccessID:  config.Public.AccessID,
		AccessKey: config.Public.AccessKey,
		Bucket:    config.Public.Bucket,
		Endpoint:  config.Public.Endpoint,
	})

	privateClient = aliyun.New(&aliyun.Config{
		AccessID:  config.Private.AccessID,
		AccessKey: config.Private.AccessKey,
		Bucket:    config.Private.Bucket,
		ACL:       aliyunoss.ACLPrivate,
		Endpoint:  config.Private.Endpoint,
	})
}

func TestAll(t *testing.T) {
	if client == nil {
		t.Skip(`skip because of no config: `)
	}

	clients := []*aliyun.Client{client, privateClient}
	for _, cli := range clients {
		tests.TestAll(cli, t)
	}
}
