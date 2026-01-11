# National Dashboard Testing Infrastructure

This directory contains the comprehensive testing infrastructure for the National Dashboard application, including end-to-end (E2E) tests, performance tests, and supporting utilities.

## Directory Structure

```
tests/
├── e2e/                          # End-to-End tests
│   ├── config/                   # Test configuration files
│   │   ├── global-setup.ts       # Global setup hook
│   │   ├── global-teardown.ts    # Global teardown hook
│   │   └── mockserver-initialization.json  # Mock server config
│   ├── pages/                    # Page Object Models
│   │   ├── LoginPage.ts          # Login page POM
│   │   ├── DashboardPage.ts      # Dashboard page POM
│   │   ├── IntelligencePage.ts   # Intelligence page POM
│   │   ├── InterventionPage.ts   # Intervention page POM
│   │   ├── CrisisConsolePage.ts  # Crisis Console page POM
│   │   └── ReportsPage.ts       
│   ├── specs # Reports page POM/                    # Test specifications
│   │   ├── intelligence-workflow.spec.ts
│   │   ├── intervention-workflow.spec.ts
│   │   ├── approval-workflow.spec.ts
│   │   └── reporting-workflow.spec.ts
│   ├── scripts/                  # Helper scripts
│   │   └── init-test-db.sql      # Database initialization
│   └── utils/                    # Test utilities
│       └── data-factory.ts       # Test data generation
├── performance/                  # Performance tests (k6)
│   ├── load-test.js              # Load testing script
│   ├── spike-test.js             # Spike testing script
│   └── soak-test.js              # Soak/endurance testing
├── playwright.config.ts          # Playwright configuration
└── README.md                     # This file
```

## Prerequisites

### System Requirements
- Node.js 18+ and npm/yarn/pnpm
- Docker and Docker Compose
- PostgreSQL 15+ (if not using Docker)
- Redis 7+ (if not using Docker)

### Installation

1. Install dependencies:
```bash
npm install
```

2. Install Playwright browsers:
```bash
npx playwright install --with-deps
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

## Environment Configuration

Create a `.env` file in the project root with the following variables:

```env
# Application
BASE_URL=http://localhost:3000
API_BASE_URL=http://localhost:3000/api/v1

# Test Users
TEST_ADMIN_EMAIL=admin@test.gov
TEST_ADMIN_PASSWORD=your_password
TEST_ANALYST_EMAIL=analyst@test.gov
TEST_ANALYST_PASSWORD=your_password
TEST_SUPERVISOR_EMAIL=supervisor@test.gov
TEST_SUPERVISOR_PASSWORD=your_password

# Database
TEST_DATABASE_URL=postgresql://test_user:test_password@localhost:5432/test_national_dashboard

# Redis
TEST_REDIS_URL=redis://localhost:6379
```

## Running Tests

### E2E Tests with Playwright

#### Run all E2E tests:
```bash
npm test:e2e
```

#### Run specific test file:
```bash
npx playwright test tests/e2e/specs/intelligence-workflow.spec.ts
```

#### Run tests with specific project:
```bash
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=mobile
```

#### Run tests in headed mode:
```bash
npx playwright test --headed
```

#### Run tests with UI:
```bash
npx playwright test --ui
```

#### Run specific test:
```bash
npx playwright test -g "should login and navigate"
```

#### Generate test report:
```bash
npx playwright show-report
```

### Performance Tests with k6

#### Run load test:
```bash
k6 run tests/performance/load-test.js
```

#### Run smoke test:
```bash
k6 run -e K6_SCRIPT=tests/performance/load-test.js --vus 1 --duration 30s
```

#### Run spike test:
```bash
k6 run tests/performance/spike-test.js --out influxdb=http://localhost:8086/k6
```

#### Run soak test:
```bash
k6 run tests/performance/soak-test.js --out json=test-results/soak-results.json
```

#### Run with environment-specific configuration:
```bash
K6_BASE_URL=http://production-api.example.com k6 run tests/performance/load-test.js
```

### Using Docker for Tests

#### Start test infrastructure:
```bash
docker-compose -f docker-compose.test.yml up -d
```

#### Run E2E tests in Docker:
```bash
docker-compose -f docker-compose.test.yml run test-e2e
```

#### Run performance tests in Docker:
```bash
docker-compose -f docker-compose.test.yml run k6-load
```

#### Stop test infrastructure:
```bash
docker-compose -f docker-compose.test.yml down
```

## Test Configuration

### Playwright Configuration

The `playwright.config.ts` file contains the following configurations:

- **Projects**: Chrome, Firefox, and Mobile Safari
- **Base URL**: Configurable via environment variable
- **Timeouts**: Page, action, and navigation timeouts
- **Retries**: Automatic retry for failed tests
- **Reporters**: HTML, JSON, and Allure reporters
- **Trace**: On-failure trace recording

### Global Setup and Teardown

- **global-setup.ts**: Initializes test database, mocks API servers, and sets up test data
- **global-teardown.ts**: Cleans up test data, closes connections, and generates reports

### Mock Server Configuration

The `mockserver-initialization.json` file configures mock API endpoints for:
- Authentication (login)
- Intelligence items
- Interventions
- Crisis alerts
- Reports

## Page Object Models

### Available POMs

| Page | Description | Key Methods |
|------|-------------|-------------|
| `LoginPage` | Authentication page | `login()`, `loginAsAdmin()`, `loginAsAnalyst()`, `getErrorMessage()` |
| `DashboardPage` | Main dashboard | `verifyDashboardStats()`, `verifyDashboardCharts()`, `getStatsCardValue()` |
| `IntelligencePage` | Intelligence management | `createIntelligence()`, `filterByStatus()`, `promoteToIntervention()` |
| `InterventionPage` | Intervention tracking | `startIntervention()`, `completeIntervention()`, `switchToKanbanView()` |
| `CrisisConsolePage` | Crisis management | `acknowledgeAlert()`, `escalateAlert()`, `centerMapOnAlert()` |
| `ReportsPage` | Report generation | `generateReport()`, `downloadReport()`, `scheduleReport()` |

## Test Specifications

### Intelligence Workflow Tests (`intelligence-workflow.spec.ts`)

Tests the complete lifecycle of intelligence items:
- Login and navigation
- Creating new intelligence items
- Filtering by status and severity
- Searching intelligence items
- Viewing item details
- Data ingestion from external sources

### Intervention Workflow Tests (`intervention-workflow.spec.ts`)

Tests the intervention management process:
- Promoting intelligence to interventions
- Starting and completing interventions
- Assigning interventions to team members
- Filtering by status and priority
- Switching between table and Kanban views

### Approval Workflow Tests (`approval-workflow.spec.ts`)

Tests multi-user approval scenarios:
- Junior analyst creates, supervisor approves
- Escalation workflows
- Role-based access control
- Multi-user concurrent access
- Workflow state transitions

### Reporting Workflow Tests (`reporting-workflow.spec.ts`)

Tests report generation and management:
- Generating weekly and monthly reports
- Creating custom reports
- Downloading in various formats (PDF, CSV)
- Sharing reports via email
- Scheduling recurring reports
- Filtering and searching reports

## Performance Testing

### Load Testing (`load-test.js`)

Simulates expected user load with configurable stages:
- **Smoke test**: Quick sanity check (1 user)
- **Load test**: Normal expected load (10-50 users)
- **Stress test**: Beyond normal capacity (50-300 users)
- **Spike test**: Sudden surge (10-100 users)
- **Soak test**: Long duration endurance (30 users, 8 hours)

### Spike Testing (`spike-test.js`)

Tests system behavior under sudden traffic spikes:
- Extreme spike scenarios
- Step spike progression
- Sudden burst testing

### Soak Testing (`soak-test.js`)

Tests system stability over extended periods:
- Memory leak detection
- Resource utilization tracking
- Error accumulation monitoring

## CI/CD Integration

### GitHub Actions Workflow

The `.github/workflows/qa.yml` file provides:
- Automated testing on push/PR
- Multi-environment testing
- Performance regression detection
- Test result reporting
- Automated deployment triggers

### Running in CI

```yaml
# Example CI step
- name: Run E2E Tests
  run: npm run test:e2e

- name: Run Performance Tests
  run: k6 run tests/performance/load-test.js
```

## Best Practices

### Writing Tests

1. **Use Page Object Models**: Encapsulate page interactions in POMs
2. **Use Data Factory**: Generate test data with `data-factory.ts`
3. **Avoid Hard-coded Selectors**: Use `data-testid` attributes
4. **Handle Async Operations**: Use proper waiting strategies
5. **Clean Up After Tests**: Use `test.afterEach()` hooks

### Test Data Management

- Use `data-factory.ts` for generating consistent test data
- Each test should create its own data
- Use unique identifiers to prevent test pollution
- Clean up test data in teardown hooks

### Performance Test Guidelines

1. **Warm up**: Include ramp-up period before measuring
2. **Think time**: Simulate realistic user delays
3. **Endpoints**: Test critical paths, not all endpoints
4. **Metrics**: Track response times, error rates, and resource usage
5. **Thresholds**: Set appropriate pass/fail criteria

## Troubleshooting

### Common Issues

#### Tests failing due to authentication
- Ensure test users exist in the database
- Check environment variables for credentials
- Verify database initialization script ran

#### Tests timing out
- Increase timeout values in `playwright.config.ts`
- Check if mock server is running
- Verify database connection

#### Performance tests showing high latency
- Check application performance metrics
- Verify database query performance
- Review network connectivity

### Debug Mode

Run tests with debug output:
```bash
DEBUG=pw:api npx playwright test
```

### Recording Traces

Enable trace recording for failed tests:
```bash
npx playwright test --trace=on
```

## Contributing

1. Follow the existing test structure
2. Add tests for new features
3. Update Page Object Models when UI changes
4. Maintain test data factories
5. Document new test scenarios
