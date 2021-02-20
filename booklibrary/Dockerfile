FROM alpine:3 as authority
RUN mkdir /user && \
    echo 'appuser:x:1000:1000:appuser:/:' > /user/passwd && \
    echo 'appgroup:x:1000:' > /user/group
RUN apk --no-cache add ca-certificates

FROM golang:1.16 as builder
WORKDIR /build

# Restore depenencies
COPY go.* ./
RUN go mod download
# Build app
COPY . .
RUN CGO_ENABLED=0 go build -o booklibrary-api ./cmd/main.go

# Set up runtime image
FROM scratch
COPY --from=authority /user/group /user/passwd /etc/
COPY --from=authority /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/booklibrary-api ./
USER appuser:appgroup
EXPOSE 5000
ENTRYPOINT ["./booklibrary-api"]