# Inno Golang HTTP Client

A simple http client made for easy communication with microservices or external API's.

## 1 Gettings started

Import the library in your project by running:

`go get github.com/inno-nl/go-http-client`

## 2 Create a new client

Call the function `httpclient.New()` to spawn a new Httpclient.

You can set defaults for requests there. The following methods are available:
```
client.ProxyUrl(string)
client.BaseUrl(string)
client.Path(string)
client.FullUrl(string)
client.Method(string)
client.Parameter(string, string)
client.Parameter(map[key]string|[]string)
client.Header(string, string)
client.Headers(map[key]value)
client.ContentType(string)
client.Body(string)
client.Json(any)
client.Timeout(int)
client.RetryCount(int)
client.BaseAuth(string, string)
client.BearerAuth(token)
```

When you want to spawn a new request, call `client.NewRequest()` which will return a pointer to a `httpclient.Request` struct.

You can use the same methods that the client has, such as `Parameter` or `Header`, to modify the request.

When you're ready to send the request, call `request.Send()`

## 3 Handling the response

The response struct has the statuscode, response headers and payload as properties.

```
response.StatusCode
response.Headers
response.Bytes()
response.String()
response.Json(pointer)
response.Xml(pointer)
response.Success()
response.Retry()
```

To unmarshal a json body to a slice, map or struct, use the following code:
```
unMarshalInto := make([]string, 0)

err := resp.Json(&unMarshalInto)
```

## 4 Error logging

The http library comes with a way to log to STDOUT. You can enable it by calling
```
client.LogErrors(true)
request.LogErrors(true)
```
