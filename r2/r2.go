package r2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/casdoor/oss"
	"io"
)

type Client struct {
	*s3.Client
	Config *Config
}

type Config struct {
	AccessID        string
	AccessKeyId     string
	AccessKeySecret string
	Bucket          string
}

func New(config *Config) *Client {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", config.AccessID),
		}, nil
	})

	client := &Client{Config: config}

	cfg, _ := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithEndpointResolverWithOptions(r2Resolver),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret, "")),
	)

	client.Client = s3.NewFromConfig(cfg)

	return client
}

// Get(path string) (*os.File, error)
func (client Client) Get(path string) (io.ReadCloser, error) {
	return nil, nil
}

// GetStream(path string) (io.ReadCloser, error)
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	return nil, nil
}

// Put(path string, reader io.Reader) (*oss.Object, error)
func (client Client) Put(path string, reader io.Reader) (*oss.Object, error) {
	return nil, nil
}

// Delete(path string) error
func Delete(path string) error {
	return nil
}

// List(path string) ([]*ossObject, error)
func List(path string) ([]*oss.Object, error) {
	return nil, nil
}

// GetURL(path string) (string, error)
func GetURL(path string) (string, error) {
	return "", nil
}

// GetEndpoint() string
func GetEndpoint() string {
	return ""
}
