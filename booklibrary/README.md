# booklibrary

## About
This is a [Go](https://golang.org/) port of [Addy Osmani's Book Library sample for
Backbone.js and Node.js](https://addyosmani.com/backbone-fundamentals/#exercise-2-book-library---your-first-restful-backbone.js-app). 

> Note: Currently, this project only contains the Web API, but not the Backbone SPA or any other client application.

## Useful commands
Here is a list of useful commands to work this sample. You can also use the included [`Task file`](https://taskfile.dev) to run them.

### Building

```bash
go build -o booklibrary-api cmd/main.go
```

### Run tests

```bash
go test -v
```

### Run app and MongoDB using Docker Compose
```bash
docker compose --profile all up -d
curl -s localhost:8000/api/books | jq
```

### Run MongoDB for local development and execute `mongosh`
```bash
docker compose up -d
docker compose exec booklibrary-db mongosh
```

### Overriding default settings for Docker Compose
You can override a few settings used by the Compose file. Either create an [`.env` file](https://docs.docker.com/compose/environment-variables/) or export them in your shell.


| Variable     | Purpose                                              | Default Value |
|--------------|------------------------------------------------------|---------------|
| `DOCKERFILE` | Dockerfile used for building the container image     | `Dockerfile`  |
| `TAG`        | Image tag used for the locally built container image | `latest`      |
| `MONGO_TAG`  | MongoDB image tag                                    | `6:0`         |



