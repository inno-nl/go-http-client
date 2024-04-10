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
r.Post("payload")
res := struct{Data any}{}
err := r.Json(&res)
```

## Agent reuse

```go
client := httpclient.New()
client.SetTimeout(60)
client.Tries = 5 // retry server errors after 1s, 2s, 4s, 8s
client.SetHeader("Accept", "application/json")

api := client.NewURL("https://localhost:8080/base?limit=100")
api.SetBasicAuth("user", "password")
api.Post(nil)
err = api.Send()
// process api.Response manually

r := api.NewURL("image") // post to "/base/image?limit=100"
r.SetHeader("Accept", "image/webp")
img, err := r.Bytes()
```

## Parameter manipulation

r := httpclient.NewURL("https://localhost")
params = url.Values{
    "config": "default",
}
params.Set("config", "override")
params.Add("limit", 10)
r.AddURL("?" + params.Encode())
r.AddURL("&debug")
