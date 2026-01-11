# NEAM Platform - Python Simulation Service

## Overview

The NEAM Simulation Service provides advanced economic simulation capabilities for what-if analysis and scenario planning. This service is part of the National Economic Activity Monitor (NEAM) platform, designed to support government decision-making through economic modeling, policy analysis, and risk assessment.

## Features

### Simulation Engines

1. **Monte Carlo Simulation**
   - Probabilistic risk assessment using random sampling
   - Support for multiple probability distributions (Normal, Uniform, Lognormal, Triangular, Exponential, Poisson, Beta, Gamma)
   - Automatic correlation handling between variables
   - Convergence diagnostics and confidence intervals

2. **Economic Projection**
   - Time-series forecasting using ARIMA and Exponential Smoothing
   - Automatic model order selection
   - Seasonality detection and modeling
   - Confidence intervals for projections

3. **Policy Scenario Analysis**
   - Compare economic outcomes under different policy interventions
   - Model fiscal policy, interest rates, tax changes, infrastructure investment
   - Multiple scenario comparison with baseline
   - Cost-effectiveness analysis

4. **Sensitivity Analysis**
   - Sobol', Morris, and FAST sensitivity indices
   - Parameter ranking by influence
   - Interaction effects analysis
   - Recommended parameter ranges

5. **Stress Testing**
   - Adverse scenario evaluation
   - Resilience scoring
   - Vulnerability identification
   - Mitigation recommendations

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/neam-platform/neam-platform.git
cd neam-platform/simulation

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
```

### Running the Service

```bash
# Development mode
python -m simulation.main

# Or using uvicorn directly
uvicorn simulation.main:app --host 0.0.0.0 --port 8083 --reload

# Production mode
uvicorn simulation.main:app --host 0.0.0.0 --port 8083 --workers 4
```

### Using Docker

```bash
# Build the image
docker build -t neam-simulation:latest .

# Run the container
docker run -p 8083:8083 \
  -e REDIS_HOST=10.112.2.4 \
  -e DATABASE_HOST=10.112.2.4 \
  neam-simulation:latest
```

## API Documentation

Once the service is running, access the interactive API documentation:

- **Swagger UI**: http://localhost:8083/docs
- **ReDoc**: http://localhost:8083/redoc

### Example Usage

#### Monte Carlo Simulation

```python
import httpx
from datetime import datetime

# Monte Carlo request
request = {
    "simulation_type": "monte_carlo",
    "simulation_id": "mc_001",
    "start_date": "2024-01-01T00:00:00",
    "end_date": "2025-01-01T00:00:00",
    "variables": [
        {
            "name": "gdp_growth",
            "description": "Annual GDP growth rate",
            "distribution": {
                "distribution_type": "normal",
                "mean": 0.025,
                "std_dev": 0.015
            },
            "unit": "percent"
        },
        {
            "name": "inflation_rate",
            "description": "Consumer price inflation",
            "distribution": {
                "distribution_type": "triangular",
                "min_value": 0.01,
                "mode": 0.02,
                "max_value": 0.05
            },
            "unit": "percent"
        }
    ],
    "output_variables": ["projected_gdp"],
    "output_formula": "100 * (1 + gdp_growth)",
    "iterations": 10000,
    "confidence_level": 0.95,
    "random_seed": 42
}

response = httpx.post(
    "http://localhost:8083/api/v1/simulations/monte-carlo",
    json=request
)
```

#### Economic Projection

```python
request = {
    "simulation_type": "economic_projection",
    "start_date": "2024-01-01T00:00:00",
    "end_date": "2024-12-01T00:00:00",
    "indicators": [
        {
            "name": "GDP",
            "code": "gdp",
            "value": 25000,
            "unit": "billion USD",
            "growth_rate": 0.02,
            "volatility": 0.01
        }
    ],
    "projection_months": 12,
    "model_type": "arima",
    "seasonality": True,
    "trend_type": "auto"
}
```

#### Policy Scenario

```python
request = {
    "simulation_type": "policy_scenario",
    "start_date": "2024-01-01T00:00:00",
    "end_date": "2026-01-01T00:00:00",
    "baseline": {
        "name": "Baseline",
        "description": "No policy intervention",
        "base_conditions": {
            "gdp": 25000,
            "inflation": 2.0,
            "unemployment": 5.0
        },
        "interventions": []
    },
    "scenarios": [
        {
            "name": "Fiscal Stimulus",
            "description": "2% GDP fiscal stimulus",
            "base_conditions": {
                "gdp": 25000,
                "inflation": 2.0,
                "unemployment": 5.0
            },
            "interventions": [
                {
                    "intervention_type": "fiscal_stimulus",
                    "name": "Infrastructure Spending",
                    "magnitude": 2.0,
                    "unit": "percent",
                    "start_date": "2024-01-01T00:00:00",
                    "lag_months": 3
                }
            ]
        }
    ],
    "indicators": ["gdp", "inflation", "unemployment"],
    "projection_months": 24,
    "compare_with_baseline": True
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis server hostname | `10.112.2.4` |
| `REDIS_PORT` | Redis server port | `6379` |
| `DATABASE_HOST` | PostgreSQL server hostname | `10.112.2.4` |
| `DATABASE_PORT` | PostgreSQL server port | `5432` |
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka brokers | `10.112.2.4:9092` |
| `CONFIG_PATH` | Path to config file | `/app/config.yaml` |

### Configuration File

See `config.yaml` for detailed configuration options including:
- Simulation engine settings
- Redis and database connections
- Kafka topics
- Logging configuration
- Performance tuning

## Architecture

### Service Components

```
simulation/
├── main.py              # FastAPI application entry point
├── config.yaml          # Configuration file
├── models/
│   ├── __init__.py
│   └── simulation.py    # Pydantic models for requests/responses
├── engines/
│   ├── __init__.py
│   ├── base.py          # Base engine class
│   ├── monte_carlo.py   # Monte Carlo simulation engine
│   ├── economic_model.py # Economic projection engine
│   ├── scenario.py      # Policy scenario engine
│   └── sensitivity.py   # Sensitivity analysis engine
├── routes/
│   ├── __init__.py
│   ├── health.py        # Health check endpoints
│   └── simulation.py    # Simulation API endpoints
└── tests/
    ├── __init__.py
    └── test_simulation.py  # Unit tests
```

### Integration with NEAM Platform

The Simulation Service integrates with other NEAM components:

- **Redis**: Caching simulation results and state management
- **PostgreSQL**: Storing historical simulation data
- **Kafka**: Event streaming for simulation events
- **Policy Engine**: Receiving policy intervention parameters
- **Intelligence Engine**: Using economic indicator data

## Testing

```bash
# Run all tests
pytest simulation/tests/ -v

# Run with coverage
pytest simulation/tests/ --cov=simulation --cov-report=html

# Run specific test
pytest simulation/tests/test_simulation.py::TestMonteCarloEngine -v
```

## Performance

### Optimization Tips

1. **Monte Carlo Iterations**: Start with 1,000 iterations, increase as needed
2. **Parallel Execution**: Use `batch` endpoint for multiple simulations
3. **Caching**: Results are automatically cached in Redis
4. **Resource Limits**: Configure based on available memory

### Expected Performance

| Operation | Typical Duration |
|-----------|-----------------|
| Monte Carlo (10k iterations) | 1-5 seconds |
| Economic Projection (24 months) | < 1 second |
| Policy Scenario (4 scenarios) | 5-30 seconds |
| Sensitivity Analysis (1000 samples) | 5-15 seconds |

## Deployment

### Kubernetes

Example deployment configuration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: neam-simulation
spec:
  replicas: 3
  selector:
    matchLabels:
      app: neam-simulation
  template:
    metadata:
      labels:
        app: neam-simulation
    spec:
      containers:
      - name: simulation
        image: neam/simulation:latest
        ports:
        - containerPort: 8083
        resources:
          limits:
            memory: "2Gi"
            cpu: "2"
          requests:
            memory: "1Gi"
            cpu: "1"
        env:
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: neam-config
              key: redis-host
```

### Docker Compose

See the main `docker-compose.yml` for full platform orchestration.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Implement the feature
5. Ensure all tests pass
6. Submit a pull request

## License

This project is proprietary software. All rights reserved.

## Support

For issues and questions:
- Create a GitHub issue
- Contact the NEAM development team
- Review the API documentation at `/docs`

## Changelog

### Version 1.0.0

- Initial release
- Monte Carlo simulation engine
- Economic projection engine
- Policy scenario analysis
- Sensitivity analysis
- Stress testing
- Comprehensive API documentation
