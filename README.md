# Golang HTTP Client

A simple http client made for easy communication with microservices or external API's.

## 1 Gettings started

Import the library in your project by running:

`go get github.com/inno-nl/go-http-client`

## 2 Create a new request

Requests are built using method chaining. Here is an example:
```
resp, err := httpclient.NewRequest().
    Url("https://inno.dev").
    Parameter("lang", "nl").
    Header("visitor_id", "AB3H71H29").
    Json(struct{Message string}{Message: "Hello, world!"}).
    Timeout(10).
    Post()
```

## 3 Handling the response

The response struct has the statuscode, response headers and payload as properties.

To unmarshal a json body to a slice, map or struct, use the following code:
```
unMarshalInto := make([]string, 0)

err := resp.Unmarshal(&unMarshalInto)
```
