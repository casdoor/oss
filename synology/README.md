# Synology

[Synology](https://www.synology.com) backend for [QOR OSS](https://github.com/qor/oss)

## Usage

```go
import "github.com/qor/oss/synology"

func main() {
  storage := synology.New(&synology.Config{
    AccessID:  "access_id",
    AccessKey: "access_key",
    Endpoint:  "your endpoint",
  })

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

