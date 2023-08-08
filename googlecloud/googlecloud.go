package googlecloud

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"github.com/casdoor/oss"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Client Google Cloud Storage
type Client struct {
	*storage.BucketHandle
	Config *Config
}

// Config Google Cloud Storage client config
type Config struct {
	AccessID     string
	AccessKey    string
	Bucket       string
	StorageClass string
	Endpoint     string
}

type Credentials struct {
	Web Web `json:"web"`
}

type Web struct {
	ClientId                string   `json:"client_id"`
	AuthUri                 string   `json:"auth_uri"`
	TokenUri                string   `json:"token_uri"`
	AuthProviderX509CertUrl string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"client_secret"`
	RedirectUris            []string `json:"redirect_uris"`
	JavascriptOrigins       []string `json:"javascript_origins"`
}

// New initialize Google Cloud Storage
func New(config *Config) (*Client, error) {
	var (
		err    error
		client = &Client{Config: config}
	)

	ctx := context.Background()

	web := Web{
		ClientId:                config.AccessID,
		AuthUri:                 "https://accounts.google.com/o/oauth2/auth",
		TokenUri:                "https://oauth2.googleapis.com/token",
		AuthProviderX509CertUrl: "https://www.googleapis.com/oauth2/v1/certs",
		ClientSecret:            config.AccessKey,
		RedirectUris:            []string{"https://www.googleapis.com/auth/cloud-platform"},
		JavascriptOrigins:       []string{"http://localhost", "https://www.googleapis.com"},
	}

	cred := Credentials{Web: web}

	credentialsData, _ := json.Marshal(cred)

	credentials, err := google.CredentialsFromJSON(context.Background(), credentialsData, "https://www.googleapis.com/auth/cloud-platform")

	if err != nil {
		log.Fatalf("Failed to get credentials: %v", err)
	}

	storageClient, err := storage.NewClient(ctx, option.WithCredentials(credentials))

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	client.BucketHandle = storageClient.Bucket(config.Bucket)

	return client, nil
}

// Get receive file with given path
func (client Client) Get(path string) (file *os.File, err error) {
	readCloser, err := client.GetStream(path)

	if err == nil {
		if file, err = ioutil.TempFile("/tmp", "googlecloud"); err == nil {
			defer readCloser.Close()
			_, err = io.Copy(file, readCloser)
			file.Seek(0, 0)
		}
	}

	return file, err
}

// GetStream get file as stream
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()
	_, err := client.BucketHandle.Object(path).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return client.BucketHandle.Object(path).NewReader(ctx)
}

// Put store a reader into given path
func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	ctx := context.Background()

	wc := client.BucketHandle.Object(urlPath).NewWriter(ctx)
	if client.Config.StorageClass != "" {
		wc.ObjectAttrs.StorageClass = client.Config.StorageClass
	}
	if _, err := io.Copy(wc, reader); err != nil {
		return nil, err
	}
	if err := wc.Close(); err != nil {
		return nil, err
	}

	attrs, err := client.BucketHandle.Object(urlPath).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return &oss.Object{
		Path:             urlPath,
		Name:             filepath.Base(urlPath),
		LastModified:     &attrs.Updated,
		StorageInterface: client,
	}, nil
}

// Delete delete file
func (client Client) Delete(path string) error {
	ctx := context.Background()
	return client.BucketHandle.Object(path).Delete(ctx)
}

// List list all objects under current path
func (client Client) List(path string) ([]*oss.Object, error) {
	var objects []*oss.Object

	ctx := context.Background()
	iter := client.BucketHandle.Objects(ctx, &storage.Query{Prefix: path})
	for {
		objAttrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		objects = append(objects, &oss.Object{
			Path:             "/" + objAttrs.Name,
			Name:             filepath.Base(objAttrs.Name),
			LastModified:     &objAttrs.Updated,
			Size:             objAttrs.Size,
			StorageInterface: client,
		})
	}

	return objects, nil
}

// GetURL get public accessible URL
func (client Client) GetURL(path string) (url string, err error) {
	return path, nil
}

func (client Client) GetEndpoint() string {
	return client.Config.Endpoint
}

func (client Client) ToRelativePath(urlPath string) string {
	if strings.HasPrefix(urlPath, client.GetEndpoint()) {
		relativePath := strings.TrimPrefix(urlPath, client.GetEndpoint())
		return strings.TrimPrefix(relativePath, "/")
	}
	return urlPath
}
