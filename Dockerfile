FROM gcr.io/distroless/static-debian11:latest
LABEL maintainer "Juan Ariza <jariza@vmware.com>"

ARG TARGETARCH
COPY api-mock_$TARGETARCH /usr/local/bin/api-mock

# by default, use a non-root (non-privileged) UID to run the container
USER 1001
EXPOSE 8080
ENV PRETTYLOG=false \
    FAILURE_RESP_BODY="{\"success\": false}" \
    FAILURE_RESP_CODE=400 \
    METHODS="GET,POST" \
    SUB_ROUTES="" \
    SUCCESS_RESP_BODY="{\"success\": true}" \
    SUCCESS_RESP_CODE=200 \
    SUCCESS_RATIO=1.0 \
    RATE_LIMIT=1000 \
    RATE_EXCEEDED_RESP_BODY="{\"success\": false, \"error\": \"rate limit exceeded\"}" \
    PORT=8080

ENTRYPOINT ["api-mock"]