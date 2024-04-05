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
r.AddURL("/json") // alter endpoint
r.AddURL("?custom=query") // keeps other parts
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

api := client.NewURL("https://localhost:8080/base?limit=100")
api.SetBasicAuth("user", "password")
api.Post(nil)
err = api.Send()
// process api.Response manually

r := api.NewPath("image") // post to "/base/image?limit=100"
r.Request.Header.Set("Accept", "image/webp")
img, err := r.Bytes()
```
