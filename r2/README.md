# Cloudflare R2

[Cloudflare R2](https://developers.cloudflare.com/r2/) 

## Usage

```go
import "https://github.com/creaplus/oss"

func main() {
  storage := r2.New(&r2.Config{{
        AccountId:       "Cloudflare AccountId",
        AccessKeyId:      "Cloudflare R2 AccessKeyId",
        AccessKeySecret:  "Cloudflare R2 AccessKeySecret",
        Bucket:          "Cloudflare R2 Bucket",
        Endpoint:        "https://{AccountId}.r2.cloudflarestorage.com",

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


