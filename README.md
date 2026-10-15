# openev-data-mcp

An [MCP](https://modelcontextprotocol.io) (Model Context Protocol) server that exposes
the [open-ev-data](https://github.com/electricitymap/open-ev-data) dataset to AI assistants and MCP clients. It lets
language models query electric vehicle specs — battery capacity, range ratings, charging capabilities, drivetrain, and
more — directly through the MCP resource protocol.

## Resources

| URI                               | Description                                                                                          |
|-----------------------------------|------------------------------------------------------------------------------------------------------|
| `evdata://vehicles`               | All electric vehicles in the dataset                                                                 |
| `evdata://makes`                  | All EV manufacturers                                                                                 |
| `evdata://vehicles/{unique_code}` | A single vehicle by unique code (`make:model:year:trim` slugs, e.g. `audi:q6_e_tron:2025:q6_e_tron`) |
| `evdata://makes/{make}/vehicles`  | All vehicles from a manufacturer (use make slug, e.g. `tesla`)                                       |

Each resource returns JSON. Vehicles are returned as records in the
[OpenEV Data schema](https://github.com/open-ev-data/open-ev-data-dataset/blob/main/schema.json): make, model, year,
trim, powertrain, battery, charging, charge ports, range, performance, dimensions, pricing, and source citations.

## Tools

| Tool              | Parameters       | Description                                                                                           |
|-------------------|------------------|-------------------------------------------------------------------------------------------------------|
| `search_vehicles` | `query` (string) | Semantic search with a natural language query, e.g. `AWD SUV with 800V fast charging and 500 km range` |

Returns the best matching vehicles (full records) ordered by similarity score. Matches with a similarity below
`0.6` (configurable) are discarded, so an empty list means nothing was relevant enough.

## Architecture

- **Transport**: Streamable HTTP (MCP port `8080`, health port `9090`)
- **Storage**: PostgreSQL — the dataset is loaded via the migration in `migrations/`
- **Caching**: In-memory TTL cache per resource URI (default `5m`, configurable)
- **Semantic search**: Qdrant collection `ev_vehicles`, embedded with Gemini (`gemini-embedding-001`). Each vehicle is
  embedded as a text document covering make, model, year, trim, variant, vehicle type, drivetrain, power, battery
  (capacity, chemistry, thermal management, heat pump), AC/DC charging power, range ratings and charging ports.

### Vector index sync

The vector index is maintained by a separate migration job, `openev-data-mcp migrate`. It provisions the
`ev_vehicles` collection if it is missing and diffs it against PostgreSQL: new vehicles are embedded and upserted,
vehicles whose indexed document changed are re-embedded, and vehicles no longer in the database are removed.
Unchanged vehicles are skipped, so re-running it makes no embedding calls unless the dataset changed (e.g. after a new
SQL migration adds vehicles). Progress is saved batch by batch, and rate-limited Gemini requests are retried, so an
interrupted run resumes where it left off.

The MCP server does not wait for the job: it serves search over whatever is already indexed.

## Running locally

```bash
GEMINI_API_KEY=... docker compose -f deployments/docker/docker-compose.yaml up
```

This starts PostgreSQL and Qdrant, runs the initial SQL migration, runs the one-shot `migrate` job to sync the vector
index, and brings up the MCP server on port `8083` (mapped from container
port `8080`). The first `migrate` run embeds every vehicle and can take several minutes on the Gemini free tier. To
re-sync after the dataset changes:

```bash
docker compose -f deployments/docker/docker-compose.yaml run --rm migrate
```

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
| `OPENEV_MCP_QDRANT_HOST`       | `localhost`  | Qdrant host         |
| `OPENEV_MCP_QDRANT_PORT`       | `6334`       | Qdrant gRPC port    |
| `OPENEV_MCP_QDRANT_API_KEY`    |              | Qdrant API key      |
| `OPENEV_MCP_QDRANT_USE_TLS`    | `false`      | Use TLS for Qdrant  |
| `OPENEV_MCP_QDRANT_COLLECTION` | `ev_vehicles` | Qdrant collection  |
| `OPENEV_MCP_QDRANT_SEARCH_LIMIT` | `10`       | Max search results  |
| `OPENEV_MCP_QDRANT_MIN_SCORE`  | `0.6`        | Minimum cosine similarity; weaker matches are discarded |
| `OPENEV_MCP_EMBEDDING_API_KEY` | (required)   | Gemini API key      |
| `OPENEV_MCP_EMBEDDING_MODEL`   | `gemini-embedding-001` | Gemini embedding model |
| `OPENEV_MCP_EMBEDDING_DIMENSIONS` | `768`     | Embedding vector size (must match the collection) |
| `OPENEV_MCP_EMBEDDING_BATCH_SIZE` | `100`     | Documents per embedding request |
| `OPENEV_MCP_REFRESH_INTERVAL`  | `5m`         | Resource cache TTL  |
| `OPENEV_MCP_PORT`              | `8080`       | MCP server port     |
| `OPENEV_MCP_HEALTH_PORT`       | `9090`       | Health check port   |

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE.md) file for details.

## Contributing

Contributions are welcome. Please read the [contributing guidelines](CONTRIBUTING.md) before opening a pull request.