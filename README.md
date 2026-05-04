# openev-data-mcp

An [MCP](https://modelcontextprotocol.io) (Model Context Protocol) server that exposes
the [open-ev-data](https://github.com/electricitymap/open-ev-data) dataset to AI assistants and MCP clients. It lets
language models query electric vehicle specs — battery capacity, range ratings, charging capabilities, drivetrain, and
more — directly through the MCP resource protocol.

## Resources

| URI                              | Description                                                    |
|----------------------------------|----------------------------------------------------------------|
| `evdata://vehicles`              | All electric vehicles in the dataset                           |
| `evdata://makes`                 | All EV manufacturers                                           |
| `evdata://vehicles/{id}`         | Full details for a single vehicle by numeric ID                |
| `evdata://makes/{make}/vehicles` | All vehicles from a manufacturer (use make slug, e.g. `tesla`) |

Each resource returns JSON. Vehicle records include make, model, year, trim, battery capacity (gross/net kWh), WLTP/EPA
range, AC/DC charging power, charge ports, drivetrain, performance figures, and source citations.

## Architecture

- **Transport**: Streamable HTTP (MCP port `8080`, health port `9090`)
- **Storage**: PostgreSQL — the dataset is loaded via the migration in `migrations/`
- **Caching**: In-memory TTL cache per resource URI (default `5m`, configurable)

## Running locally

```bash
docker compose -f deployments/docker/docker-compose.yaml up
```

This starts PostgreSQL, runs the initial migration, and brings up the MCP server on port `8083` (mapped from container
port `8080`).

## Configuration

All settings can be set via environment variables (prefix `OPENEV_MCP_`) or a config file passed with `--config`.

| Env var                        | Default      | Description         |
|--------------------------------|--------------|---------------------|
| `OPENEV_MCP_DATABASE_HOST`     | `localhost`  | PostgreSQL host     |
| `OPENEV_MCP_DATABASE_PORT`     | `5432`       | PostgreSQL port     |
| `OPENEV_MCP_DATABASE_USER`     | `openevdata` | Database user       |
| `OPENEV_MCP_DATABASE_PASSWORD` | `openevdata` | Database password   |
| `OPENEV_MCP_DATABASE_DBNAME`   | `openevdata` | Database name       |
| `OPENEV_MCP_DATABASE_SSL_MODE` | `disable`    | PostgreSQL SSL mode |
| `OPENEV_MCP_REFRESH_INTERVAL`  | `5m`         | Resource cache TTL  |
| `OPENEV_MCP_PORT`              | `8080`       | MCP server port     |
| `OPENEV_MCP_HEALTH_PORT`       | `9090`       | Health check port   |

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE.md) file for details.

## Contributing

Contributions are welcome. Please read the [contributing guidelines](CONTRIBUTING.md) before opening a pull request.