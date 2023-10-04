# booklibrary

## About
This is a [Go](https://golang.org/) port of [Addy Osmani's Book Library sample for
Backbone.js and Node.js](https://addyosmani.com/backbone-fundamentals/#exercise-2-book-library---your-first-restful-backbone.js-app). It uses
[slog](https://pkg.go.dev/log/slog) and exposes a Prometheus endpoint to scrape metrics.

> Note: This project only contains the Web API, but not the original Backbone SPA or any other client application.

## Useful commands
Here is a list of useful commands to work this sample. You can also use the included [`Task file`](./Taskfile.yml) to run them if you have [Task](https://taskfile.dev) installed.

### Building

```bash
go build -o booklibrary-api cmd/booklibrary-api/main.go

# Using Task
task build
```
### Running (without explicit build)
```bash
go run cmd/booklibrary-api/main.go

# Using Task
task run
```



### Runing tests

```bash
go test -v ./...

# Using Task
task test
```

### Build Docker image
```bash
docker buildx build -t booklibrary-api --load .

# Using Task
task docker:build
```

### Run app and MongoDB using Docker Compose
```bash
docker compose --profile all up -d
curl -s localhost:8000/api/books | jq
# Shutdown
docker compose --profile all down


# Using Task
task docker:up
curl -s localhost:8000/api/books | jq
# Shutdown
task docker:down
```

### Run MongoDB for local development and execute `mongosh`
```bash
docker compose up -d
docker compose exec booklibrary-db mongosh
# Shutdown
docker compose down

# Using Task
task mongo:up
task mongo:sh
# Shutdown
task mongo:down

```

### Overriding default settings for Docker Compose
You can override a few settings used by the Compose file. Either create an [`.env` file](https://docs.docker.com/compose/environment-variables/) or export them in your shell.


| Variable     | Purpose                                              | Default Value |
|--------------|------------------------------------------------------|---------------|
| `DOCKERFILE` | Dockerfile used for building the container image     | `Dockerfile`  |
| `TAG`        | Image tag used for the locally built container image | `latest`      |
| `MONGO_TAG`  | MongoDB image tag                                    | `6:0`         |



