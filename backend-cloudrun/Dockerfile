# Build stage: ultra-fast, zero-dependency static compilation
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# Compile static binaries with optimizations stripped (-s -w)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o job ./cmd/job

# Final stage: minimal distroless image (~18MB, near-instant sub-10ms cold start)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /app/server /server
COPY --from=builder /app/job /job

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
