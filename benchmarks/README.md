# UUID benchmarks

`go test ./benchmarks -run '^$' -bench '^BenchmarkUUIDGeneration$' -benchtime=1s -count=3` compares random UUIDv4 and UUIDv7 generation, including entropy acquisition, without database calls.

For the database benchmark, use a dedicated disposable PostgreSQL database, for example:

```sh
docker run --rm -d --name uuid-bench -e POSTGRES_PASSWORD=benchmark -e POSTGRES_DB=benchmark -p 127.0.0.1:55432:5432 postgres:17-alpine
UUID_BENCH_DSN='host=127.0.0.1 port=55432 user=postgres password=benchmark dbname=benchmark sslmode=disable' go test ./benchmarks -run '^$' -bench '^BenchmarkPostgresUUID$' -benchtime=1x -count=3
docker stop uuid-bench
```

The benchmark creates a randomly named table and drops it. Each run creates 100,000 IDs before timing, inserts rows through GORM’s generic API into an initially empty table with a `TEXT` primary key and a 22-byte payload, in batches of 1,000 with GORM's default write transaction and `PrepareStmt: false` (matching the application). It measures the index size and 1,000 deterministic random successful primary-key lookups after the inserts; these are warm-cache lookups. Values are averages per row or lookup. Use `-benchtime=1x`; standard Go ns/op is intentionally suppressed. Index metrics include the primary-key index only. The client and database are on the same machine; results do not include realistic network latency, concurrent traffic, existing rows, or cold cache. PostgreSQL text indexes use the database's collation, unlike native UUID columns. An all-new v7 table does not measure a mixed existing database.

One local run on an Intel i7-12650H, Go 1.27.4, GORM 1.31.2, PostgreSQL 17.11 Alpine, database collation `en_US.utf8` produced the following. The production Neon server is PostgreSQL 17.11 on ARM64 (aarch64, Debian GNU/Linux); the local benchmark used PostgreSQL 17.11 on x86-64 Alpine. The development compose file specifies PostgreSQL 16. These local measurements do not represent Neon performance: hardware, cloud environment, network latency, and workload differ. Concurrent machine load affected the run; repeat under controlled conditions before drawing performance conclusions.

| Version | Generation ns/op (3 runs) | Insert µs/row (3 runs) | Lookup µs/op (3 runs) | Index bytes (3 runs) |
| --- | --- | --- | --- | --- |
| v4 | 85.29, 84.42, 83.94 | 7.279, 9.697, 5.327 | 360.5, 233.5, 142.9 | 7,880,704; 7,667,712; 7,643,136 |
| v7 | 158.3, 147.7, 156.3 | 3.501, 3.349, 3.291 | 97.55, 124.6, 91.08 | 5,931,008; 5,931,008; 5,931,008 |

These measurements suggest smaller v7 indexes in this particular fresh-table run, while insert and lookup times were noisy. UUIDv7 does not guarantee faster lookups, and existing UUIDv4 values remain unchanged.

References: [GORM hooks](https://gorm.io/docs/hooks.html), [GORM performance](https://gorm.io/docs/performance.html), [GORM data types](https://gorm.io/docs/data_types.html), [google/uuid v1.6.0 NewV7](https://pkg.go.dev/github.com/google/uuid@v1.6.0#NewV7), [uuidjs v7](https://github.com/uuidjs/uuid#uuidv7options-buffer-offset), and [RFC 9562](https://www.rfc-editor.org/rfc/rfc9562.html).
