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

package oss

import (
	"io"
	"os"
	"time"
)

// StorageInterface define common API to operate storage
type StorageInterface interface {
	Get(path string) (*os.File, error)
	GetStream(path string) (io.ReadCloser, error)
	Put(path string, reader io.Reader) (*Object, error)
	Delete(path string) error
	List(path string) ([]*Object, error)
	GetURL(path string) (string, error)
	GetEndpoint() string
}

// Object content object
type Object struct {
	Path             string
	Name             string
	LastModified     *time.Time
	Size             int64
	StorageInterface StorageInterface
}

// Get retrieve object's content
func (object Object) Get() (*os.File, error) {
	return object.StorageInterface.Get(object.Path)
}
