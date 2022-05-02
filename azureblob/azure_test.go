// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package azureblob

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/qor/oss/tests"
)

var client *Client

func init() {
	client = New(&Config{
		AccessId:  "",
		AccessKey: "",
		Bucket:    "",
		Region:    "",
		Endpoint:  "localhost:8080",
	})
}

func TestClientPut(t *testing.T) {
	f, err := ioutil.ReadFile("C:\\Users\\MI\\Pictures\\Wallpaper-1.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = client.Put("test.png", bytes.NewReader(f))
	if err != nil {
		return
	}
}

func TestClientPut2(t *testing.T) {
	tests.TestAll(client, t)
}

func TestClientDelete(t *testing.T) {
	fmt.Println(client.Delete("test.png"))
}
