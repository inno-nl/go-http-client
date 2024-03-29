# Inno Golang HTTP Client

A simple http client made for easy communication with microservices or external API's.

## Simple request

```go
import "github.com/inno-nl/go-http-client"
data, err := httpclient.New("http://localhost/test").Bytes()
```
