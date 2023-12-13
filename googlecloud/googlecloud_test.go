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
	"github.com/casdoor/oss/googlecloud"
	"io/ioutil"
	"testing"
)

var client *googlecloud.Client

func TestInit(t *testing.T) {
	client, _ = googlecloud.New(&googlecloud.Config{
		AccessID:  "",
		AccessKey: "",
		Bucket:    "",
		Endpoint:  "",
	})

}

func TestClientPut(t *testing.T) {
	f, err := ioutil.ReadFile("D:\\1.txt")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = client.Put("test.txt", bytes.NewReader(f))
	if err != nil {
		return
	}
}

func TestClientDelete(t *testing.T) {
	fmt.Println(client.Delete("test.png"))
}

func TestClientList(t *testing.T) {
	fmt.Println(client.List("/"))
}

func TestClientGet(t *testing.T) {
	fmt.Println(client.Get("/"))
}
