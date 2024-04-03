# Inno Golang HTTP Client

A simple http client made for easy communication with microservices or external API's.

## Simple request

```go
import "github.com/inno-nl/go-http-client"
data, err := httpclient.New("http://localhost/test").Bytes()
```

## Extended request

```go
r := httpclient.New("https://httpbin.org/status/404")
r.Timeout = 60
r.URL.Path = "/json" // alter endpoint
r.Parameters.Set("custom", "query")
r.Request.Header.Del("User-Agent")
r.Post("payload")
r.Tries = 5 // retry server errors after 1s, 2s, 4s, 8s
res := struct{Data any}{}
err := r.Json(&res)
```
