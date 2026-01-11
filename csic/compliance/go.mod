module github.com/csic-platform/compliance

go 1.21

require (
	github.com/csic-platform/shared v0.0.0
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.5.0
	github.com/lib/pq v1.10.9
	go.uber.org/zap v1.26.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/csic-platform/shared => ../shared
