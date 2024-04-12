# Inno Golang HTTP Client

A simple http client made for easy communication with microservices or external API's.

## Direct download

```go
import "github.com/inno-nl/go-http-client"
data, err := httpclient.NewURL("http://localhost/test").Text()
```

## Extended request

```go
r := httpclient.NewURL("https://httpbin.org/status/404")
r.AddURL("/json") // alter endpoint
r.AddURL("?custom=query") // keeps other parts
r.Post("payload")
res := struct{Data any}{}
err := r.Json(&res)
```

## Agent reuse

```go
client := httpclient.New()
client.SetTimeout(60)
client.SetRetry(4) // retry server errors after 1s, 2s, 4s, 8s
client.SetHeader("Accept", "application/json")

api := client.NewURL("https://localhost:8080/base")
api.AddQuery("limit", 1)
api.SetBasicAuth("user", "password")
api.Post(nil)
err = api.Send()
// process api.Response manually

r := api.NewURL("image") // post to "/base/image?limit=1"
r.AddQuery("greeting", "hello?") // .AddURL("&greeting=hello%3f")
r.SetHeader("Accept", "image/webp")
img, err := r.Bytes()
```
