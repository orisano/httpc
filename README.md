# httpc
[![Build Status](https://travis-ci.org/orisano/httpc.svg?branch=master)](https://travis-ci.org/orisano/httpc)
[![Maintainability](https://api.codeclimate.com/v1/badges/2c91b8e3d8b367c2400c/maintainability)](https://codeclimate.com/github/orisano/httpc/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/2c91b8e3d8b367c2400c/test_coverage)](https://codeclimate.com/github/orisano/httpc/test_coverage)

## How to Use
```go
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/orisano/httpc"
)

type User struct {
	ID       string
	Password string
	Age      int
}

func main() {
	rb, err := httpc.NewRequestBuilder("http://api.example/", nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	client := http.DefaultClient

	user := &User{
		ID:       "john",
		Password: "foobar",
		Age:      28,
	}

	req, err := rb.NewRequest(ctx, http.MethodPost, "/v1/users", httpc.WithJSON(user))
	if err != nil {
		log.Fatal(err)
	}
	resp, err := httpc.Retry(client, req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
```

## Author
Nao Yonashiro (@orisano)

## License
MIT
