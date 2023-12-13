# OSS

[![Go Report Card](https://goreportcard.com/badge/github.com/casdoor/oss)](https://goreportcard.com/report/github.com/casdoor/oss)
[![Go](https://github.com/casdoor/oss/actions/workflows/ci.yml/badge.svg)](https://github.com/casdoor/oss/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/casdoor/oss.svg)](https://pkg.go.dev/github.com/casdoor/oss)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/casdoor/oss)

Casdoor OSS aims to provide a common interface to operate files with any kinds of storages, like cloud storages, FTP, file system etc

- [Local File System](https://github.com/casdoor/oss/tree/master/filesystem)
- [MinIO (Open Source)](https://min.io)
- [AWS S3](https://aws.amazon.com/s3)
- [Azure Blob Storage](https://azure.microsoft.com/en-us/products/storage/blobs)
- [Google Cloud - Cloud Storage](https://cloud.google.com/storage)
- [Alibaba Cloud OSS & CDN](https://cn.aliyun.com/product/oss)
- [Tencent Cloud COS](https://cloud.tencent.com/product/cos)
- [Qiniu Cloud Kodo](https://www.qiniu.com/products/kodo)

# Usage

## Installation

```
git clone https://github.com/casdoor/oss
```

## Create Client

Different oss providers need to provide different configuration, but we support a unit API as below to create the oss client.

```go
func New(config *Config) (*Client, error)
```

The config generally includes the following information:

- `AccessID`: The access ID for authentication.
- `AccessKey`: The access key for authentication.
- `Bucket`: The name of the bucket where the data is stored.
- `Endpoint`: The endpoint for accessing the storage service.

Please note that the actual configuration may vary depending on the specific storage service being used.

## Operation

Currently, QOR OSS provides support for file system, S3, Aliyun and so on, You can easily implement your own storage strategies by implementing the interface.

```go
type StorageInterface interface {
  Get(path string) (*os.File, error)
  GetStream(path string) (io.ReadCloser, error)
  Put(path string, reader io.Reader) (*Object, error)
  Delete(path string) error
  List(path string) ([]*Object, error)
  GetEndpoint() string
  GetURL(path string) (string, error)
}
```

## Example

Here's an example of how to use [QOR OSS](https://github.com/qor/oss) with S3. After initializing the s3 storage, The functions in the interface are available.

```go
import (
  "github.com/oss/filesystem"
  "github.com/oss/s3"
  awss3 "github.com/aws/aws-sdk-go/s3"
)

func main() {
  storage := s3.New(s3.Config{AccessID: "access_id", AccessKey: "access_key", Region: "region", Bucket: "bucket", Endpoint: "cdn.getqor.com", ACL: awss3.BucketCannedACLPublicRead})
  // storage := filesystem.New("/tmp")

  // Save a reader interface into storage
  storage.Put("/sample.txt", reader)

  // Get file with path
  storage.Get("/sample.txt")

  // Get object as io.ReadCloser
  storage.GetStream("/sample.txt")

  // Delete file with path
  storage.Delete("/sample.txt")

  // List all objects under path
  storage.List("/")

  // Get Public Accessible URL (useful if current file saved privately)
  storage.GetURL("/sample.txt")
}
```

# License

Released under the [MIT License](http://opensource.org/licenses/MIT).
