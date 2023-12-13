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
	client, err := googlecloud.New(&googlecloud.Config{
		AccessID:  "xxx",
		AccessKey: "-----BEGIN PRIVATE KEY-----xxx\\n-----END PRIVATE KEY-----\\n",
		Bucket:    "casdoor",
		Endpoint:  "https://storage.googleapis.com",
	})
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
	fmt.Println(client.Delete("123.txt"))
}

func TestClientList(t *testing.T) {
	client := getClient()
	fmt.Println(client.List("/"))
}

func TestClientGet(t *testing.T) {
	client := getClient()
	fmt.Println(client.Get("/"))
}
