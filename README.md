# wait-database-cli

CLI tool to wait for MySQL / PostgreSQL server to be up and running

## Features

- Supports both MySQL and PostgreSQL databases
- Configurable via command-line flags or environment variables
- Optional query execution after connection is established
- Timeout configuration (default: 10 minutes)
- Support for SOCKS proxy (MySQL only)

## Installation

```bash
go install github.com/evalphobia/wait-database-cli@latest
```

## Usage

### Command Line Flags

- `-type`: Database type (`mysql` or `postgresql`, default: `mysql`)
- `-host`: Database host
- `-port`: Database port (default: 3306 for MySQL, 5432 for PostgreSQL)
- `-user`: Database user (default: `root` for MySQL, `postgres` for PostgreSQL)
- `-password`: Database password
- `-database`: Database name
- `-query`: SQL query to execute after connection

### Environment Variables

- `WAIT_DATABASE_TYPE`: Database type
- `WAIT_DATABASE_HOST`: Database host
- `WAIT_DATABASE_PORT`: Database port
- `WAIT_DATABASE_USER`: Database user
- `WAIT_DATABASE_PASSWORD`: Database password
- `WAIT_DATABASE_DATABASE`: Database name
- `WAIT_DATABASE_QUERY`: SQL query to execute

### Examples

#### MySQL

```bash
# Using command line flags
wait-database-cli -type mysql -host localhost -port 3306 -user root -password mypassword -database mydb

# Using environment variables
export WAIT_DATABASE_TYPE=mysql
export WAIT_DATABASE_HOST=localhost
export WAIT_DATABASE_USER=root
export WAIT_DATABASE_PASSWORD=mypassword
export WAIT_DATABASE_DATABASE=mydb
wait-database-cli

# With query execution
wait-database-cli -type mysql -host localhost -database mydb -query "SELECT VERSION()"
```

#### PostgreSQL

```bash
# Using command line flags
wait-database-cli -type postgresql -host localhost -port 5432 -user postgres -password mypassword -database mydb

# Using environment variables
export WAIT_DATABASE_TYPE=postgresql
export WAIT_DATABASE_HOST=localhost
export WAIT_DATABASE_USER=postgres
export WAIT_DATABASE_PASSWORD=mypassword
export WAIT_DATABASE_DATABASE=mydb
wait-database-cli

# With query execution
wait-database-cli -type postgresql -host localhost -database mydb -query "SELECT version()"
```

## Docker

The tool is available as a Docker image:

```bash
docker run -e WAIT_DATABASE_TYPE=mysql -e WAIT_DATABASE_HOST=127.0.0.1 -e WAIT_DATABASE_USER=root -e WAIT_DATABASE_PASSWORD=pass --rm evalphobia/wait-database-cli
```

## License

MIT
