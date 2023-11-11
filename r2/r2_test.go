package r2_test

import (
	"fmt"
	"github.com/casdoor/oss/r2"
	"github.com/casdoor/oss/tests"
	"github.com/jinzhu/configor"
	"testing"
)

type Config struct {
	AccountId       string `env:"CF_R2_Account_ID"`
	AccessKeyId     string `env:"CF_R2_ACCESS_KEY_ID"`
	AccessKeySecret string `env:"CF_R2_ACCESS_KEY_SECRET"`
	Bucket          string `env:"CF_R2_BUCKET"`
	Endpoint        string `env:"CF_R2_ENDPOINT"`
}

var (
	client *r2.Client
	config = Config{}
)

func init() {
	configor.Load(&config)

	client = r2.New(&r2.Config{
		AccountId:       config.AccountId,
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		Bucket:          config.Bucket,
		Endpoint:        config.Endpoint,
	})

}

func TestAll(t *testing.T) {
	fmt.Println("testing r2 with object public")
	tests.TestAll(client, t)
	TestToRelativePath(t)
}

func TestToRelativePath(t *testing.T) {
	urlMap := map[string]string{
		"https://mybucket.s3.amazonaws.com/myobject.ext": "myobject.ext",
		"https://qor-example.com/myobject.ext":           "myobject.ext",
		"//mybucket.s3.amazonaws.com/myobject.ext":       "myobject.ext",
		"http://mybucket.s3.amazonaws.com/myobject.ext":  "myobject.ext",
		"myobject.ext": "myobject.ext",
	}

	for url, path := range urlMap {
		if client.ToRelativePath(url) != path {
			t.Errorf("%v's relative path should be %v, but got %v", url, path, client.ToRelativePath(url))
		}
	}
}
