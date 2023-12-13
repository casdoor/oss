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

package googlecloud_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/casdoor/oss/googlecloud"
)

func getClient() *googlecloud.Client {
	serviceAccountJson := `{
  "type": "service_account",
  "project_id": "casbin",
  "private_key_id": "xxx",
  "private_key": "-----BEGIN PRIVATE KEY-----\nxxx\n-----END PRIVATE KEY-----\n",
  "client_email": "casdoor-service-account@casbin.iam.gserviceaccount.com",
  "client_id": "10336152244758146xxx",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/casdoor-service-account%40casbin.iam.gserviceaccount.com",
  "universe_domain": "googleapis.com"
}`

	config := &googlecloud.Config{
		ServiceAccountJson: serviceAccountJson,
		Bucket:             "casdoor",
		Endpoint:           "",
	}

	client, err := googlecloud.New(config)
	if err != nil {
		panic(err)
	}

	return client
}

func TestClientPut(t *testing.T) {
	f, err := ioutil.ReadFile("E:/123.txt")
	if err != nil {
		panic(err)
	}

	client := getClient()
	_, err = client.Put("123.txt", bytes.NewReader(f))
	if err != nil {
		panic(err)
	}
}

func TestClientDelete(t *testing.T) {
	client := getClient()
	err := client.Delete("123.txt")
	if err != nil {
		panic(err)
	}
}

func TestClientList(t *testing.T) {
	client := getClient()
	objects, err := client.List("/")
	if err != nil {
		panic(err)
	}

	fmt.Println(objects)
}

func TestClientGet(t *testing.T) {
	client := getClient()
	f, err := client.Get("/")
	if err != nil {
		panic(err)
	}

	fmt.Println(f)
}
