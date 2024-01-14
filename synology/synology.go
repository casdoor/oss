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

package synology

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"io/ioutil"

	"mime/multipart"
	"github.com/casdoor/oss"
)

// Client Synology NAS storage
type Client struct {
	Config        *Config
	SID       string
	SynoToken string
	AppAPIList map[string]map[string]interface{}
	FullAPIList map[string]map[string]interface{}
}

// Config Synology NAS client config
type Config struct {
	Endpoint string
	AccessID string
	AccessKey string
	SessionExpire bool
	Verify   bool
	Debug   bool
	OtpCode string
	SharedFolder string
}

func New(config *Config) *Client {
	client := &Client{Config: config}
	client.Login("FileStation")
	client.GetAPIList("FileStation")
	return client
}


// Get receive file with given path
func (client Client) Get(path string) (file *os.File, err error) {
	readCloser, err := client.GetStream(path)
	if err != nil {
		return nil, err
	}

	if file, err = ioutil.TempFile("/tmp", "synology"); err == nil {
		defer readCloser.Close()
		_, err = io.Copy(file, readCloser)
		file.Seek(0, 0)
	}

	return file, err
}

// GetStream get file as stream
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	sharedFolder := client.Config.SharedFolder
	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	apiName := "SYNO.FileStation.Download"


	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "download")
	params.Set("path", sharedFolder + path)
	params.Set("mode", "download")
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SID)

	url := baseURL + "?" + params.Encode()
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SID)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed, status code: %d", resp.StatusCode)
	}

	return resp.Body, err
}


func (client *Client) GetAPIList(app string) error {
	baseURL := client.Config.Endpoint + "/webapi/"
	queryPath := "query.cgi?api=SYNO.API.Info"
	params := url.Values{}
	params.Set("version", "1")
	params.Set("method", "query")
	params.Set("query", "all")
	
	response, err := http.Get(baseURL + queryPath + "&" + params.Encode())

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return err
	}

	var responseJSON map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&responseJSON)

	if err != nil {
		return err
	}
	responseJSONTwoLevel := make(map[string]map[string]interface{})
	for key, value := range responseJSON["data"].(map[string]interface{}) {
		if innerMap, ok := value.(map[string]interface{}); ok {
			responseJSONTwoLevel[key] = innerMap
		}
	}

	client.AppAPIList = make(map[string]map[string]interface{})
	if app != "" {
		for key := range responseJSONTwoLevel {
			if strings.Contains(strings.ToLower(key), strings.ToLower(app)) {
				client.AppAPIList[key] = responseJSONTwoLevel[key]
			}
		}
	} else {
		client.FullAPIList = responseJSONTwoLevel
	}

	return nil
}


func (client *Client) Login(application string) error {
	baseURL := client.Config.Endpoint + "/webapi/"
	loginAPI := "auth.cgi?api=SYNO.API.Auth"
	params := url.Values{}
	params.Set("version", "3")
	params.Set("method", "login")
	params.Set("account", client.Config.AccessID)
	params.Set("passwd", client.Config.AccessKey)
	params.Set("session", application)
	params.Set("format", "cookie")
	params.Set("enable_syno_token", "yes")

	if client.Config.OtpCode != "" {
		params.Set("opt_code", client.Config.OtpCode)
	}
	loginAPI = loginAPI + "&" + params.Encode()

	var sessionRequestJSON map[string]interface{}
	if !client.Config.SessionExpire && client.SID != "" {
		client.Config.SessionExpire = false
		if client.Config.Debug {
			fmt.Println("User already logged in")
		}
	} else {
		// Check request for error:
		response, err := http.Get(baseURL + loginAPI)
		if err != nil {
			return err
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return err
		}

		err = json.NewDecoder(response.Body).Decode(&sessionRequestJSON)
		if err != nil {
			return err
		}
	}
	
	// Check DSM response for error:
	errorCode := client.getErrorCode(sessionRequestJSON)

	if errorCode == 0 {
		client.SID = sessionRequestJSON["data"].(map[string]interface{})["sid"].(string)
		client.SynoToken = sessionRequestJSON["data"].(map[string]interface{})["synotoken"].(string)
		client.Config.SessionExpire = false
		if client.Config.Debug {
			fmt.Println("User logged in, new session started!")
		}
	} else {
		client.SID = ""
		if client.Config.Debug {
			fmt.Println("User logged faild")
		}
	}

	return nil

}

func (client Client) getErrorCode(response map[string]interface{}) int {

	var code int
	if response["success"].(bool) {
		code = 0 // No error
	} else {
		errorData := response["error"].(map[string]interface{})
		code = int(errorData["code"].(float64))
	}

	return code
}

func (client *Client) Put(urlPath string, reader io.Reader) (r *oss.Object, err error) {
	sharedFolder := client.Config.SharedFolder

	apiName := "SYNO.FileStation.Upload"
	baseURL := client.Config.Endpoint + "/webapi/"
	loginAPI := "entry.cgi"

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "upload")
	params.Set("SynoToken", client.SynoToken)

	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	parserURL, err := url.Parse(urlPath)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
	}
	path := parserURL.Path
	dir := filepath.Dir(path)
	// change windows path to linux path
	dir = filepath.ToSlash(dir)

	err = writer.WriteField("path", sharedFolder + dir)
	if err != nil {
		return nil, err
	}
	
	err = writer.WriteField("overwrite", "true")
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("create_parents", "true")
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(urlPath)
	part, err := writer.CreateFormFile("file", filename) // Set a placeholder filename
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	url := baseURL + loginAPI + "?" + params.Encode()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SID)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary
	
	resp, err := http.DefaultClient.Do(req)
	
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed, status code: %d", resp.StatusCode)
	}

	now := time.Now()
	return &oss.Object{
		Path: urlPath,
		Name: filepath.Base(urlPath),
		LastModified: &now,
		StorageInterface: client,
	}, nil

}

// Delete delete file
func (client Client) Delete(path string) error {
	sharedFolder := client.Config.SharedFolder

	apiName := "SYNO.FileStation.Delete"

	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "start")
	params.Set("path", sharedFolder + path)
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SID)

	req_url := baseURL + "?" + params.Encode()
	
	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SID)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary

	resp, err := http.Get(req_url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	var responseJSON map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseJSON)
	if err != nil {
		return err
	}

	return nil
}

// List list all objects under current path
func (client Client) List(path string) (objects []*oss.Object, err error) {
	sharedFolder := client.Config.SharedFolder

	apiName := "SYNO.FileStation.List"

	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "list")
	params.Set("folder_path", sharedFolder + "/" + path)
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SID)

	req_url := baseURL + "?" + params.Encode()
	
	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SID)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary

	resp, err := http.Get(req_url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var responseJSON map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseJSON)
	if err != nil {
		return nil, err
	}

	for _, content := range responseJSON["data"].(map[string]interface{})["files"].([]interface{}) {
		now := time.Now()
		path := content.(map[string]interface{})["path"].(string)
		// remove top shared path
		parsedUrl, err := url.Parse(path)
		if err != nil {
			return nil, err
		}
		pathParts := strings.Split(parsedUrl.Path, "/")
		if len(pathParts) > 1 {
			pathParts = append(pathParts[:1], pathParts[2:]...)
		}
		parsedUrl.Path = strings.Join(pathParts, "/")
		path = parsedUrl.String()

		objects = append(objects, &oss.Object{
			Path:             path,
			Name:             filepath.Base(content.(map[string]interface{})["path"].(string)),
			LastModified:     &now,
			StorageInterface: &client,
		})
	}

	return objects, err
}


// GetEndpoint get endpoint, FileSystem's endpoint is /
func (client Client) GetEndpoint() string {
	return client.Config.Endpoint
}

// GetURL get public accessible URL
func (client Client) GetURL(path string) (get_url string, err error) {
	sharedFolder := client.Config.SharedFolder
	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	// get file stream
	apiName := "SYNO.FileStation.Download"

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "download")
	params.Set("path", sharedFolder + path)
	params.Set("mode", "download")
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SID)

	get_url = baseURL + "?" + params.Encode()

	return get_url, nil
}
