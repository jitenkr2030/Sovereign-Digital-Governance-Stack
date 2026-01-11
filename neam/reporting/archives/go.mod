module neam-platform/reporting/archives

go 1.21

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.17.1
	github.com/gin-gonic/gin v1.9.1
	github.com/redis/go-redis/v9 v9.3.0
	neam-platform/shared v0.0.0
)

replace neam-platform/shared => ../shared
