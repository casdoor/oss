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

package googlecloud

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/casdoor/oss"
)

// Client Google Cloud Storage
type Client struct {
	Config       *Config
	BucketHandle *storage.BucketHandle
}

// Config Google Cloud Storage client config
type Config struct {
	ServiceAccountJson string
	Bucket             string
	Endpoint           string
}

// New initializes Google Cloud Storage
func New(config *Config) (*Client, error) {
	var (
		ctx   = context.Background()
		scope = "https://www.googleapis.com/auth/cloud-platform"

		credentials *google.Credentials
		err         error
	)

	if config.ServiceAccountJson != "" {
		credentials, err = google.CredentialsFromJSON(ctx, []byte(config.ServiceAccountJson), scope)
	} else {
		credentials, err = google.FindDefaultCredentials(ctx, scope)
	}
	if err != nil {
		return nil, err
	}

	storageClient, err := storage.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, err
	}

	client := &Client{
		Config:       config,
		BucketHandle: storageClient.Bucket(config.Bucket),
	}
	return client, nil
}

// Get receives file with given path
func (client Client) Get(path string) (file *os.File, err error) {
	readCloser, err := client.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()

	file, err = ioutil.TempFile("/tmp", "googlecloud")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, readCloser)
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GetStream gets file as stream
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	ctx := context.Background()
	_, err := client.BucketHandle.Object(path).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return client.BucketHandle.Object(path).NewReader(ctx)
}

// Put stores a reader into given path
func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	ctx := context.Background()

	wc := client.BucketHandle.Object(urlPath).NewWriter(ctx)

	_, err := io.Copy(wc, reader)
	if err != nil {
		return nil, err
	}

	err = wc.Close()
	if err != nil {
		return nil, err
	}

	attrs, err := client.BucketHandle.Object(urlPath).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	res := &oss.Object{
		Path:             urlPath,
		Name:             filepath.Base(urlPath),
		LastModified:     &attrs.Updated,
		StorageInterface: client,
	}
	return res, nil
}

// Delete deletes file
func (client Client) Delete(path string) error {
	ctx := context.Background()
	return client.BucketHandle.Object(path).Delete(ctx)
}

// List lists all objects under current path
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
	if client.Config.Endpoint != "" {
		return client.Config.Endpoint
	}
	return "https://storage.googleapis.com"
}

func (client Client) ToRelativePath(urlPath string) string {
	if strings.HasPrefix(urlPath, client.GetEndpoint()) {
		relativePath := strings.TrimPrefix(urlPath, client.GetEndpoint())
		return strings.TrimPrefix(relativePath, "/")
	}
	return urlPath
}
