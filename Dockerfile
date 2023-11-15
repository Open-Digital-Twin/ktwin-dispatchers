# Build stage
FROM golang:1.20 as builder
ARG SERVICE_NAME

WORKDIR /workspace

# Copy Go files
COPY go.mod go.mod
COPY go.sum go.sum

# Copy source
COPY cmd cmd
COPY pkg pkg

RUN CGO_ENABLED=0 go build -a -o service cmd/$SERVICE_NAME/main.go

# Run image
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/service .

EXPOSE 8080

CMD ["./service"]
