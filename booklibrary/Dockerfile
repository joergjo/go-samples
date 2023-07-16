FROM --platform=$BUILDPLATFORM golang:1.20 AS build
WORKDIR /src
ARG TARGETOS TARGETARCH
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /out/ ./cmd/booklibrary-api

FROM --platform=$BUILDPLATFORM gcr.io/distroless/static:nonroot AS final
EXPOSE 8000
COPY --from=build /out/booklibrary-api ./
ENTRYPOINT ["./booklibrary-api"]
