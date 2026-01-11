module neam-platform/black-economy

go 1.21

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.17.1
	github.com/gin-gonic/gin v1.9.1
	github.com/neo4j/neo4j-go-driver/v5 v5.12.0
	github.com/redis/go-redis/v9 v9.3.0
	neam-platform/shared v0.0.0
)

replace neam-platform/shared => ../shared
