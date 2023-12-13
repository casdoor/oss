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

package tencent

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/casdoor/oss/tests"
)

func TestClient_Get(t *testing.T) {

}

var client *Client

func init() {
	client = New(&Config{
		AppID:     "1252882253",
		AccessID:  "AKIDToxukQWBG8nGXcBN8i662nOo12sc5Wjl",
		AccessKey: "40jNrBf5mLiuuiU8HH7lDTXP5at00sbA",
		Bucket:    "tets-1252882253",
		Region:    "ap-shanghai",
		ACL:       "public-read", // private，public-read-write，public-read；默认值：private
		//Endpoint:  config.Public.Endpoint,
	})
}

func TestClient_Put(t *testing.T) {
	f, err := ioutil.ReadFile("/home/owen/Downloads/2.png")
	if err != nil {
		t.Error(err)
		return
	}

	client.Put("test.png", bytes.NewReader(f))
}

func TestClient_Put2(t *testing.T) {
	tests.TestAll(client, t)
}

func TestClient_Delete(t *testing.T) {
	fmt.Println(client.Delete("test.png"))
}
