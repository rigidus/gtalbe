# Testing

The project includes unit and integration tests to ensure correctness of implementation. To run all tests:

```sh
go test ./...


## Run tests

Tests use Testcontainers to create isolated PostgreSQL containers. Make sure Docker is installed and running.

### Requirements
- Docker and Docker Compose
- Go 1.20 or higher (to run locally)

### Running Tests via Docker Compose
1. Create an `.env` file with environment variables (see `.env.example` for an example).
2. Run the command:
   ```sh
 docker-compose -f docker-compose.yml run --rm -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal app go test ./....
