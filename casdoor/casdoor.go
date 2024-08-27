// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

package casdoor

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/casdoor/oss"
)

type Client struct {
	*casdoorsdk.Client
	Config       *Config
	Prefix       string
	CustomDomain string
	httpClient   *http.Client
}

type Config struct {
	AccessID         string
	AccessKey        string
	Endpoint         string
	Certificate      string
	ApplicationName  string
	OrganizationName string
	Provider         string
}

func New(config *Config) *Client {
	casdoorClient := casdoorsdk.NewClient(config.Endpoint, config.AccessID, config.AccessKey, config.Certificate, config.OrganizationName, config.ApplicationName)
	client := &Client{
		casdoorClient,
		config,
		"",
		"",
		&http.Client{},
	}
	provider, err := client.GetProvider(config.Provider)
	if err != nil {
		panic(err)
	}
	client.Prefix = path.Join(provider.Bucket, provider.PathPrefix)
	client.CustomDomain = provider.Domain
	if strings.HasSuffix(client.CustomDomain, "/") {
		client.CustomDomain = client.CustomDomain[:len(client.CustomDomain)-1]
	}
	return client
}

func (client Client) Get(path string) (file *os.File, err error) {
	readCloser, err := client.GetStream(path)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(readCloser)

	if file, err = os.CreateTemp(os.TempDir(), "casdoor"); err != nil {
		defer readCloser.Close()
		_, err = io.Copy(file, readCloser)
		file.Seek(0, 0)
	}

	return file, err

}

func (client Client) GetStream(path string) (io.ReadCloser, error) {
	path, err := client.GetURL(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	//defer func(Body io.ReadCloser) {
	//	err := Body.Close()
	//	if err != nil {
	//		return
	//	}
	//}(resp.Body)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		return nil, fmt.Errorf("%s", string(respBytes))
	}

	return resp.Body, nil
}

func (client Client) Put(urlPath string, reader io.Reader) (r *oss.Object, err error) {
	if seeker, ok := reader.(io.ReadSeeker); ok {
		seeker.Seek(0, 0)
	}

	var buffer []byte
	buffer, err = io.ReadAll(reader)

	//urlPath = client.transUrl(urlPath)

	fileUrl, name, err := client.UploadResource("casdoor-oss", "", "", client.transUrl(urlPath), buffer)

	now := time.Now()
	return &oss.Object{
		Path:             fileUrl,
		Name:             name,
		LastModified:     &now,
		StorageInterface: client,
	}, err
}

func (client Client) Delete(path string) error {
	name, err := client.getName(path)
	if err != nil {
		return err
	}
	_, err = client.DeleteResource(&casdoorsdk.Resource{Application: client.ApplicationName, Provider: client.Config.Provider, Name: name})
	return err
}

func (client Client) List(rawPath string) ([]*oss.Object, error) {
	var objects []*oss.Object
	rawPath, err := client.getName(rawPath)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(rawPath, "/") {
		rawPath = rawPath[1:]
	}

	resourceList, err := client.GetResources(client.Config.OrganizationName, "casdoor-oss", "provider", client.Config.Provider, "Direct", rawPath)
	if err != nil {
		return nil, err
	}

	for _, item := range resourceList {
		t, err := time.Parse(time.RFC3339, item.CreatedTime)
		if err != nil {
			return nil, err
		}

		objects = append(objects, &oss.Object{
			Path:             item.Url,
			Name:             item.Name,
			LastModified:     &t,
			StorageInterface: client,
		})
	}

	return objects, nil
}

func (client Client) GetEndpoint() string {
	return client.Config.Endpoint
}

func (client Client) getName(rawPath string) (string, error) {
	urlPath, err := url.Parse(client.CustomDomain)
	if err != nil {
		return "", err
	}
	return path.Join(urlPath.Path, client.transUrl(rawPath)), nil
}

func (client Client) GetURL(path string) (url string, err error) {
	return client.CustomDomain + client.transUrl(path), nil
}

func (client Client) transUrl(urlPath string) string {
	return strings.Replace(urlPath, "\\", "/", -1)
}
