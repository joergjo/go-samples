FROM golang:1.15 as builder
WORKDIR /build

# Restore depenencies firs
COPY go.mod .
COPY go.sum .
RUN go mod download
# Build app
COPY . .
RUN go build ./cmd/main.go

# Set up runtime image
FROM alpine:3 AS final
RUN apk --no-cache add ca-certificates
EXPOSE 5000
RUN adduser -S -D -H -h /app appuser
USER appuser
WORKDIR /app
COPY --from=builder /build/main .
ENTRYPOINT ["./main"]