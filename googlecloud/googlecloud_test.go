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
