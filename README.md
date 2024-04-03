# Inno Golang HTTP Client

A simple http client made for easy communication with microservices or external API's.

## Direct download

```go
import "github.com/inno-nl/go-http-client"
data, err := httpclient.New("http://localhost/test").String()
```

## Extended request

```go
r := httpclient.New("https://httpbin.org/status/404")
r.URL.Path = "/json" // alter endpoint
r.Parameters.Set("custom", "query")
r.Request.Header.Del("User-Agent")
r.Post("payload")
res := struct{Data any}{}
err := r.Json(&res)
```

## Agent reuse

```go
client := httpclient.New("https://httpbin.org")
client.Timeout(60)
client.Tries = 5 // retry server errors after 1s, 2s, 4s, 8s
client.Request.Header.Set("Accept", "application/json")

r1 := client.Clone()
r1.URL.Path = "/image"
r1.Request.Header.Set("Accept", "image/webp")
img, err := r1.Bytes()

r2 := client.Clone()
r2.URL.Path = "/status/500"
err = r2.Send()
```
