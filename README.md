# Inno Golang HTTP Client

A simple http client made for easy communication with microservices or external API's.

## Direct download

```go
import "github.com/inno-nl/go-http-client"
data, err := httpclient.NewURL("http://localhost/test").String()
```

## Extended request

```go
r := httpclient.NewURL("https://httpbin.org/status/404")
r.Request.URL.Path = "/json" // alter endpoint
r.Parameters.Set("custom", "query")
r.Request.Header.Del("User-Agent")
r.Post("payload")
res := struct{Data any}{}
err := r.Json(&res)
```

## Agent reuse

```go
client := httpclient.New()
client.SetTimeout(60)
client.Tries = 5 // retry server errors after 1s, 2s, 4s, 8s
client.Request.Header.Set("Accept", "application/json")

r1 := client.NewURL("https://httpbin.org/status/500")
err = r1.Send()

r2 := r1.NewPath("/image")
r2.Request.Header.Set("Accept", "image/webp")
img, err := r2.Bytes()
```
