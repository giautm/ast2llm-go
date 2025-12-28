# syntax=docker/dockerfile:1.6
FROM gcr.io/distroless/static:nonroot

ARG TARGETPLATFORM
ARG BINARY=ast2llm-go

COPY ${TARGETPLATFORM}/${BINARY} /usr/local/bin/${BINARY}

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/ast2llm-go"]
