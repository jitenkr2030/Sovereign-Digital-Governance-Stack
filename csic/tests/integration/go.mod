module github.com/csic-platform/tests/integration

go 1.21

require (
	github.com/csic-platform/services/mining v0.0.0
	github.com/csic-platform/services/compliance v0.0.0
	github.com/csic-platform/services/transaction-monitoring v0.0.0
	github.com/csic-platform/services/energy-integration v0.0.0
	github.com/testcontainers/testcontainers-go v0.26.0
	github.com/testcontainers/postgres v0.0.0
	github.com/stretchr/testify v1.8.4
	github.com/google/uuid v1.4.0
	github.com/jackc/pgx/v5 v5.5.1
)

replace github.com/csic-platform/services/mining => ../services/mining
replace github.com/csic-platform/services/compliance => ../services/compliance
replace github.com/csic-platform/services/transaction-monitoring => ../services/transaction-monitoring
replace github.com/csic-platform/services/energy-integration => ../services/energy-integration
