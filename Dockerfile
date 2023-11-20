FROM gcr.io/distroless/static-debian11:latest
LABEL maintainer "Juan Ariza <jariza@vmware.com>"

ARG TARGETARCH
COPY dist/api-mock_linux_${TARGETARCH}*/api-mock /usr/local/bin/

# by default, use a non-root (non-privileged) UID to run the container
USER 1001
EXPOSE 8080
ENV API_TOKEN="" \
    LOG_LEVEL="info" \
    FAILURE_RESP_BODY="" \
    FAILURE_RESP_CODE=400 \
    METHODS="GET,POST" \
    RESP_DELAY=0 \
    SUB_ROUTES="" \
    SUCCESS_RESP_BODY="" \
    SUCCESS_RESP_CODE=200 \
    SUCCESS_RATIO=1.0 \
    RATE_LIMIT=1000 \
    RATE_EXCEEDED_RESP_BODY="" \
    PORT=8080

ENTRYPOINT ["api-mock"]
