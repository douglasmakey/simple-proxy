# Simple proxy

A simple implementation proxy in go.

## Dependencies
* go >= 1.12
* gorilla/mux

## Run

```sh
go run main.go
```

## Run with Docker
```sh
docker build -t IMAGE_NAME .
docker run -it -p 8000:8000 IMAGE_NAME
```