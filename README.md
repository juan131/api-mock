# REST API mock

A Golang service that can be customized to mock REST APIs

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Usage](#usage)
- [Configuration](#configuration)
- [Build](#build)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Usage

Run the API mock using Docker:

```bash
docker run --rm -p 8080:8080 juanariza131/api-mock
```

The API mock will be available at `http://localhost:8080/v1/mock`.

## Configuration

The API mock can be configured with the following environment variables:

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `PORT` | The port to listen on | `8080` |
| `LOG_LEVEL` | The log level | `info` |
| `API_TOKEN` | Bearer token to authenticate requests | `` |
| `FAILURE_RESP_BODY` | The response body to return when mocking a failure | `{"error":{"message":"failed request","code":1005,"id":"[random-value]"}}` |
| `FAILURE_RESP_CODE` | The HTTP status code to return when mocking a failure | `400` |
| `SUCCESS_RESP_BODY` | The response body to return when mocking a success | `{"success": "true"}` |
| `SUCCESS_RESP_CODE` | The HTTP status code to return when mocking a success | `200` |
| `SUCCESS_RATIO` | The ratio of success to failure responses | `1.0` |
| `METHODS` | The HTTP methods to mock | `GET,POST` |
| `RESP_DELAY` | The response delay (in milliseconds) | `0` |
| `SUB_ROUTES` | The sub routes to mock | `` |
| `RATE_LIMIT` | The API rate limit (requests per second) | `1000` |
| `RATE_EXCEEDED_RESP_BODY` | The response body to return when mocking a rate exceeded | `{"error":{"message":"rate limit exceeded","code":1004,"id":"[random-value]"}}` |

## Build

You can build the API mock binary with the following command:

```bash
make build
```

The binary will be available in the `out` directory.
