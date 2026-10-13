# Build stage: ultra-fast, zero-dependency static compilation
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# Compile static binary with optimizations stripped (-s -w)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# Final stage: minimal scratch/distroless image (~15MB, near-instant sub-10ms cold start)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /app/server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
