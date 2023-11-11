package r2

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/casdoor/oss"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	R2     *s3.Client
	Config *Config
}

type Config struct {
	AccountId       string
	AccessKeyId     string
	AccessKeySecret string
	Bucket          string
	Endpoint        string
}

// New init cloudflare r2 store
func New(config *Config) *Client {

	client := &Client{Config: config}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: config.Endpoint,
		}, nil
	})

	cfg, _ := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithEndpointResolverWithOptions(r2Resolver),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret, "")),
	)

	client.R2 = s3.NewFromConfig(cfg)

	return client
}

// Get file with path
func (client Client) Get(path string) (file *os.File, err error) {
	stream, err := client.GetStream(path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("r2*%s", ext)

	if err == nil {
		if file, err = ioutil.TempFile("/tmp", pattern); err == nil {
			defer stream.Close()
			_, err = io.Copy(file, stream)
			file.Seek(0, 0)
		}
	}

	return file, err
}

// GetStream Get object as io.ReadCloser
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(client.Config.Bucket),
		Key:    aws.String(client.ToRelativePath(path)),
	}
	object, err := client.R2.GetObject(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	return object.Body, nil
}

// Put Save a reader interface into storage
func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	if seeker, ok := reader.(io.ReadSeeker); ok {
		seeker.Seek(0, 0)
	}

	urlPath = client.ToRelativePath(urlPath)
	buffer, err := ioutil.ReadAll(reader)

	fileType := mime.TypeByExtension(path.Ext(urlPath))
	if fileType == "" {
		fileType = http.DetectContentType(buffer)
	}

	params := &s3.PutObjectInput{
		Bucket:        aws.String(client.Config.Bucket), // required
		Key:           aws.String(urlPath),              // required
		Body:          bytes.NewReader(buffer),
		ContentLength: int64(len(buffer)),
		ContentType:   aws.String(fileType),
	}

	_, err = client.R2.PutObject(context.TODO(), params)

	now := time.Now()
	return &oss.Object{
		Path:             urlPath,
		Name:             filepath.Base(urlPath),
		LastModified:     &now,
		StorageInterface: client,
	}, err
}

// Delete file with path
// Deprecated: Feature Not Implemented; https://developers.cloudflare.com/r2/api/s3/api/
func (client Client) Delete(path string) error {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(client.Config.Bucket),
		Key:    aws.String(client.ToRelativePath(path)),
	}

	_, err := client.R2.DeleteObject(context.TODO(), params)
	if err != nil {
		return err
	}
	return nil
}

// List all objects under path
func (client Client) List(path string) ([]*oss.Object, error) {
	var objects []*oss.Object
	var prefix string

	if path != "" {
		prefix = client.ToRelativePath(path)
	}

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(client.Config.Bucket),
		Prefix: aws.String(prefix),
	}
	listObjectsResponse, err := client.R2.ListObjectsV2(context.TODO(), params)
	if err == nil {
		for _, content := range listObjectsResponse.Contents {
			objects = append(objects, &oss.Object{
				Path:             client.ToRelativePath(*content.Key),
				Name:             filepath.Base(*content.Key),
				LastModified:     content.LastModified,
				StorageInterface: client,
			})
		}
	}

	return objects, err
}

// GetURL Public Accessible URL (useful if current file saved privately)
// Deprecated: Feature Not Implemented; https://developers.cloudflare.com/r2/api/s3/api/
func (client Client) GetURL(path string) (string, error) {
	return "/" + client.ToRelativePath(path), nil
}

// GetEndpoint string
func (client Client) GetEndpoint() string {
	if client.Config.Bucket == "" {
		endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", client.Config.AccountId)
		client.Config.Bucket = endpoint
	}

	return client.Config.Endpoint
}

var urlRegexp = regexp.MustCompile(`(https?:)?//((\w+).)+(\w+)/`)

// ToRelativePath process path to relative path
func (client Client) ToRelativePath(urlPath string) string {
	if urlRegexp.MatchString(urlPath) {
		if u, err := url.Parse(urlPath); err == nil {
			return strings.TrimPrefix(u.Path, "/")
		}
	}

	return strings.TrimPrefix(urlPath, "/")
}
