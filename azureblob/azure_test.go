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
