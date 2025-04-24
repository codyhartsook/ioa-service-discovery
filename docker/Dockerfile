# Get cross platform build tools
FROM --platform=$BUILDPLATFORM tonistiigi/xx AS xx

FROM --platform=$BUILDPLATFORM golang:1.23 AS builder

# Import cross platform build tools to build environment
COPY --from=xx / /

WORKDIR /workspace/

# Copy the Go Modules manifests
COPY ./go.mod /workspace/
COPY ./go.sum /workspace/

RUN go mod download

# Copy the go source files
COPY ./cmd /workspace/cmd
COPY ./internal /workspace/internal
COPY ./pkg /workspace/pkg

# Build
ARG TARGETOS TARGETARCH TARGETPLATFORM
ENV CGO_ENABLED=0
RUN xx-go build -o bin/main ./cmd

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/bin/main .
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/main"]