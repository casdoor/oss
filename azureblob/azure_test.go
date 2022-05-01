package azureblob

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/qor/oss/tests"
)

func TestClient_Get(t *testing.T) {

}

var client *Client

func init() {
	client = New(&Config{
		AccessID:  "AKIDToxukQWBG8nGXcBN8i662nOo12sc5Wjl",
		AccessKey: "40jNrBf5mLiuuiU8HH7lDTXP5at00sbA",
		Bucket:    "tets-1252882253",
		Region:    "ap-shanghai",
		Endpoint:  "localhost:8080",
	})
}

func TestClient_Put(t *testing.T) {
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

func TestClient_Put2(t *testing.T) {
	tests.TestAll(client, t)
}

func TestClient_Delete(t *testing.T) {
	fmt.Println(client.Delete("test.png"))
}
