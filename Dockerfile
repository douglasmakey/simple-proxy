FROM golang:latest  as builder
ENV GO111MODULE=on
WORKDIR /go/src/github.com/douglasmakey/simple-proxy
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=builder go/src/github.com/douglasmakey/simple-proxy/app .
CMD ["./app"]