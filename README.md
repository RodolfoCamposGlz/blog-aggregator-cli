# Gator CLI

Gator CLI is a command-line tool for managing and scraping RSS feeds. This README will guide you through setting up the project, installing dependencies, configuring the program, and running commands.

## Prerequisites

Before you can run the program, you need to have the following installed:

- **PostgreSQL**: A relational database to store data like feeds and posts.
- **Go (1.18 or higher)**: Go is the programming language used to build the application.

You can download and install these from their respective websites:

- [PostgreSQL](https://www.postgresql.org/download/)
- [Go](https://go.dev/dl/)

Additionally, `Gator` uses the following tools for database management:

- [sqlc](https://github.com/kyleconroy/sqlc): A tool for generating Go code from SQL queries. This is used for type-safe queries with PostgreSQL.
- [Goose](https://github.com/pressly/goose): A tool for database migrations, which is used to manage database schema changes.

### Install `gator` CLI using `go install`

To install the `gator` CLI, you'll need to have Go installed on your machine. If Go is installed, you can easily install the `gator` CLI by running the following command in your terminal:

```bash
go install https://github.com/RodolfoCamposGlz/blog-aggregator-cli@latest

### Update gatorconfig.json

Update the value db_url with your respective local postgres configuration
```
