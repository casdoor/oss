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
	"context"
	"fmt"

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

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/casdoor/oss"
)

// Client azure blob storage
type Client struct {
	Config       *Config
	containerURL *azblob.ContainerURL
}

type Config struct {
	AccessId  string //Account Name
	AccessKey string //Access Keys
	Region    string
	Bucket    string //Container Name
	Endpoint  string //endpoint
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

const blobFormatString = `https://%s.blob.core.windows.net`

var (
	ctx = context.Background() // This uses a never-expiring context.
)

func New(config *Config) *Client {
	var client = &Client{Config: config}

	serviceURL, _ := GetBlobService(config)
	client.containerURL = containerUrl(serviceURL, config)
	return client
}

func GetBlobService(config *Config) (azblob.ServiceURL, error) {
	// Use your Storage account's name and key to create a credential object; this is used to access your account.
	credential, err := azblob.NewSharedKeyCredential(config.AccessId, config.AccessKey)
	if err != nil {
		return azblob.ServiceURL{}, err
	}

	// Create a request pipeline that is used to process HTTP(S) requests and responses. It requires
	// your account credentials. In more advanced scenarios, you can configure telemetry, retry policies,
	// logging, and other options. Also, you can configure multiple request pipelines for different scenarios.
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your Storage account blob service URL endpoint.
	// The URL typically looks like this:
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, config.AccessId))

	// Create an ServiceURL object that wraps the service URL and a request pipeline.
	return azblob.NewServiceURL(*u, p), nil
}

func containerUrl(serviceURL azblob.ServiceURL, config *Config) *azblob.ContainerURL {
	// This returns a ContainerURL object that wraps the container's URL and a request pipeline (inherited from serviceURL)
	container := serviceURL.NewContainerURL(config.Bucket)
	return &container
}

func (client Client) UploadBlob(blobName *string, blobType *string, data io.ReadSeeker) (azblob.BlockBlobURL, error) {
	// Create a URL that references a to-be-created blob in your Azure Storage account's container.
	// This returns a BlockBlobURL object that wraps the blob's URL and a request pipeline (inherited from containerUrl)
	blobURL := client.containerURL.NewBlockBlobURL(*blobName) // Blob names can be mixed case

	// Upload the blob
	_, err := blobURL.Upload(ctx, data, azblob.BlobHTTPHeaders{ContentType: *blobType}, azblob.Metadata{}, azblob.BlobAccessConditions{}, azblob.DefaultAccessTier, nil, azblob.ClientProvidedKeyOptions{}, azblob.ImmutabilityPolicyOptions{})
	if err != nil {
		return azblob.BlockBlobURL{}, err
	}

	return blobURL, nil
}

func (client Client) DownloadBlob(blobName *string) (*azblob.DownloadResponse, error) {
	// Create a URL that references a to-be-created blob in your Azure Storage account's container.
	// This returns a BlockBlobURL object that wraps the blob's URL and a request pipeline (inherited from containerUrl)
	blobURL := client.containerURL.NewBlockBlobURL(*blobName) // Blob names can be mixed case

	// Download the blob's contents and verify that it worked correctly
	return blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
}

func (client Client) DeleteBlob(blobName *string) error {
	// Create a URL that references a to-be-created blob in your Azure Storage account's container.
	// This returns a BlockBlobURL object that wraps the blob's URL and a request pipeline (inherited from containerUrl)
	blobURL := client.containerURL.NewBlockBlobURL(*blobName) // Blob names can be mixed case

	// Delete the blob
	_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		return err
	}

	return nil
}

func (client Client) GetListBlob() ([][]azblob.BlobItemInternal, error) {
	var results [][]azblob.BlobItemInternal

	// List the blob(s) in our container; since a container may hold millions of blobs, this is done 1 segment at a time.
	for marker := (azblob.Marker{}); marker.NotDone(); { // The parens around Marker{} are required to avoid compiler error.
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := client.containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return nil, err
		}
		// IMPORTANT: ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			fmt.Print("Blob name: " + blobInfo.Name + "\n")
		}

		results = append(results, listBlob.Segment.BlobItems)
	}

	return results, nil
}

func (client Client) Get(path string) (file *os.File, err error) {
	path = client.ToRelativePath(path)
	readCloser, err := client.GetStream(path)

	if err == nil {
		if file, err = ioutil.TempFile("/tmp", "ali"); err == nil {
			defer func(readCloser io.ReadCloser) {
				err := readCloser.Close()
				if err != nil {

				}
			}(readCloser)
			_, err = io.Copy(file, readCloser)
			_, err := file.Seek(0, 0)
			if err != nil {
				return nil, err
			}
		}
	}
	return file, err
}

func (client Client) GetStream(path string) (io.ReadCloser, error) {
	name := path
	blob, err := client.DownloadBlob(&name)
	if err != nil {
		return nil, err
	}
	return blob.Response().Body, err
}

func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	if seeker, ok := reader.(io.ReadSeeker); ok {
		_, err := seeker.Seek(0, 0)
		if err != nil {
			return nil, err
		}
	}
	urlPath = client.ToRelativePath(urlPath)
	buffer, err := ioutil.ReadAll(reader)

	fileType := mime.TypeByExtension(path.Ext(urlPath))
	if fileType == "" {
		fileType = http.DetectContentType(buffer)
	}

	if fileType == "" {
		fileType = http.DetectContentType(buffer)
	}

	_, err = client.UploadBlob(&urlPath, &fileType, bytes.NewReader(buffer))
	if err != nil {
		return nil, err
	}
	now := time.Now()

	return &oss.Object{
		Path:             urlPath,
		Name:             filepath.Base(urlPath),
		LastModified:     &now,
		StorageInterface: client,
	}, err
}

func (client Client) Delete(path string) error {
	path = client.ToRelativePath(path)
	return client.DeleteBlob(&path)
}

func (client Client) List(path string) ([]*oss.Object, error) {
	panic("implement me")
}

func (client Client) GetURL(path string) (string, error) {
	return path, nil
}

func (client Client) GetEndpoint() string {
	if client.Config.Endpoint != "" {
		return client.Config.Endpoint
	}
	return client.containerURL.String()
}
