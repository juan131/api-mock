# REST API mock

A Golang service that can be customized to mock REST APIs

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Usage](#usage)
- [Configuration](#configuration)
- [Build](#build)
  - [Container build](#container-build)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Usage

Run the API mock using Docker:

```bash
docker run --rm -p 8080:8080 juanariza131/api-mock
```

The API mock will be available at `http://localhost:8080`.

## Configuration

The API mock can be configured with the following environment variables:

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `PORT` | The port to listen on | `8080` |
| `FAILURE_RESP_BODY` | The response body to return when mocking a failure | `{"success": "false"}` |
| `FAILURE_RESP_CODE` | The HTTP status code to return when mocking a failure | `400` |
| `SUCCESS_RESP_BODY` | The response body to return when mocking a success | `{"success": "true"}` |
| `SUCCESS_RESP_CODE` | The HTTP status code to return when mocking a success | `200` |
| `SUCCESS_RATIO` | The ratio of success to failure responses | `1.0` |
| `METHODS` | The HTTP methods to mock | `GET,POST` |
| `SUB_ROUTES` | The sub routes to mock | `` |
| `RATE_LIMIT` | The API rate limit (requests per second) | `1000` |
| `RATE_EXCEEDED_RESP_BODY` | The response body to return when mocking a rate exceeded | `{"success": "false", "error": "rate limit exceeded"}` |

## Build

You can build the API mock binary with the following command:

```bash
make build
```

The binary will be available in the `out` directory.

### Container build

To build this container with multi-arch support, run the commands below:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-mock_amd64 ./cmd/api-mock
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o api-mock_arm64 ./cmd/api-mock
docker buildx create --name image-builder
docker buildx use image-builder
docker buildx build . --platform linux/amd64,linux/arm64 -t api-mock
docker buildx rm image-builder
```
