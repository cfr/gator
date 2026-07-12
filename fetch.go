package main

import (
    "net/http"
    "errors"
    "io"
    "context"
    "encoding/xml"
)

func get[T any](ctx context.Context, url string) (*T, error) {
    data, err := getData(ctx, url)
    if (err != nil) {
        return nil, err
    }
    return parse[T](data)
}

func parse[T any](data []byte) (*T, error) {
    var v T
    err := xml.Unmarshal(data, &v)
    return &v, err
}

func getData(ctx context.Context, url string) ([]byte, error) {

    var data []byte

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return data, err
    }
    req.Header.Set("User-Agent", "gator")
    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        return data, err
    }
    body, err := io.ReadAll(res.Body)
    res.Body.Close()
    if res.StatusCode > 299 {
        return data, errors.New("Response failed with status code")
    }
    if err != nil {
        return data, err
    }
    data = body
    return data, nil
}

