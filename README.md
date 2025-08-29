# Translation Service

[![CI](https://github.com/henok321/translation-service/actions/workflows/CI.yml/badge.svg)](https://github.com/henok321/translation-service/actions/workflows/CI.yml)
[![Deploy](https://github.com/henok321/translation-service/actions/workflows/deploy.yml/badge.svg)](https://github.com/henok321/translation-service/actions/workflows/deploy.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=henok321_translation-service&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=henok321_translation-service)

## Synopsis

A golang-based translation service.

## CI/CD

The project uses GitHub Actions for CI/CD. The [CI Workflow](.github/workflows/CI.yml) runs on push for the main branch
and for pull requests.
The [CD workflow](.github/workflows/deploy.yml) runs on push to the main branch.

## Database Migration

The project uses `goose` for database migrations. Migrations are located in the `db_migration` directory.

## Prerequisites

Ensure the following dependencies are installed:

- [Go](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/get-docker/)
- [pre-commit](https://pre-commit.com/) (`pip install pre-commit`)
- [Goose](https://github.com/pressly/goose) (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Setup and Development

### Run Setup

Execute the following command to set up the project:

```sh
make setup
```

This command will:

- Install commit hooks.
- Start the local database.
- Run database migrations.
- Create a `.env` file with necessary environment variables.

Reset database:

```shell
make reset
```

### Start the Application

To run the application locally:

```shell
set -o allexport
source .env
set +o allexport
go run cmd/main.go
```

### Build and run binary

#### Build

```shell
make build
```

#### Run

```shell
set -o allexport
source .env
set +o allexport
./translation-service
```

### Makefile targets

For more information on available Makefile targets, run:

```shell
make help
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
