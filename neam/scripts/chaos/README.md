# Chaos Engineering Implementation for NEAM Platform

This directory contains the complete implementation of chaos engineering capabilities for the NEAM Platform, enabling systematic testing of system resilience under failure conditions.

## Quick Start

### 1. Install Prerequisites

Ensure you have the following tools installed:
- kubectl (v1.18+)
- helm (v3.2.1+)
- python3 (v3.8+)
- jq (optional, for formatting)

Install Python dependencies:
```bash
cd scripts/chaos
pip install -r requirements.txt
```

### 2. Deploy Chaos Mesh

Deploy Chaos Mesh to your Kubernetes cluster:
```bash
cd scripts/chaos
./chaos-operator.sh deploy
```

Verify the installation:
```bash
./chaos-operator.sh verify
```

### 3. Apply Chaos Experiments

Apply predefined chaos experiments:
```bash
./chaos-operator.sh apply ../../k8s/chaos/experiments.yaml
```

Or use templates for common scenarios:
```bash
./chaos-operator.sh apply ../../k8s/chaos/templates.yaml
```

### 4. Run Validation

Run data loss validation and performance degradation tests:
```bash
./chaos-operator.sh validate --all
```

## Directory Structure

```
scripts/chaos/
├── chaos-operator.sh           # Main operations script
├── data_loss_validator.py      # Data consistency validation
├── performance_degradation_test.py  # Performance testing
├── requirements.txt            # Python dependencies
└── test_config.json           # Test configuration

k8s/chaos/
├── experiments.yaml            # Chaos experiment manifests
├── templates.yaml             # Reusable experiment templates
└── test_config.json           # Test configuration reference

docs/chaos-engineering/
├── README.md                   # Documentation overview
├── CHAOS-ENGINEERING-GUIDE.md  # Comprehensive guide
├── INFRASTRUCTURE-EXPERIMENTS.md  # Infrastructure experiments
├── APPLICATION-EXPERIMENTS.md  # Application experiments
├── RESILIENCE-VALIDATION.md    # Data validation procedures
└── PERFORMANCE-TESTING.md      # Performance testing guide
```

## Key Features

### Infrastructure Experiments
- **Pod Chaos**: Pod failure, container kill, pod restart
- **Network Chaos**: Latency, packet loss, network partition, bandwidth limits
- **Stress Chaos**: CPU, memory, and I/O stress injection

### Application Experiments
- **HTTP Chaos**: Latency injection, error injection, abort simulation
- **Service Dependency**: External service failure simulation

### Validation Framework
- **Data Loss Validator**: Multi-store consistency checking (PostgreSQL, MongoDB, Redis)
- **Performance Testing**: Load generation with integrated chaos injection
- **Metrics Collection**: Latency, throughput, error rates, resource utilization

## Usage Examples

### Deploy and Run a Single Experiment
```bash
# Deploy Chaos Mesh
./chaos-operator.sh deploy

# Apply experiments
./chaos-operator.sh apply ../../k8s/chaos/experiments.yaml

# Monitor specific experiment
./chaos-operator.sh monitor pod-kill-frontend

# Check experiment status
./chaos-operator.sh status pod-kill-frontend

# Clean up when done
./chaos-operator.sh delete pod-kill-frontend
```

### Run Performance Degradation Test
```bash
# Run baseline test
python3 performance_degradation_test.py \
  --base-url http://api-gateway:8080 \
  --rps 100 \
  --duration 300 \
  --chaos-type latency \
  --target api-service \
  --output results.json
```

### Validate Data Consistency
```bash
# Configure data stores in config.yaml
# Run validation
python3 data_loss_validator.py \
  --config config.yaml \
  --phase all \
  --duration 300
```

### Use Templates
```bash
# Apply specific template
kubectl apply -f ../../k8s/chaos/templates.yaml -l template=critical-path

# Run health check workflow
kubectl apply -f ../../k8s/chaos/templates.yaml -l template=health-check
```

## Configuration

### Chaos Operator Options
```bash
./chaos-operator.sh --help

Commands:
  deploy          Deploy Chaos Mesh
  verify          Verify installation
  apply [file]    Apply experiments
  list            List experiments
  status [name]   Check status
  delete [name]   Delete experiments
  monitor [name]  Monitor execution
  validate        Run validation
  report          Generate report
  cleanup         Clean up all
  test            Run test suite
```

### Performance Test Configuration
Edit `k8s/chaos/test_config.json` to customize:
- Load patterns (RPS, concurrency, duration)
- Chaos scenarios (types, targets, parameters)
- Thresholds (latency, throughput, error rates)
- Data stores for validation

## Safety Guidelines

1. **Start Small**: Begin with non-critical environments
2. **Set Blast Radius**: Limit experiments to specific namespaces or labels
3. **Monitor Closely**: Watch metrics during experiment execution
4. **Have Abort Ready**: Know how to stop experiments quickly
5. **Communicate**: Notify relevant teams before running experiments

### Recommended Thresholds
- Maximum error rate: 50%
- Maximum p99 latency: 5000ms
- Maximum CPU utilization: 95%
- Auto-abort on critical conditions: enabled

## Best Practices

### Experiment Design
- Define clear hypotheses before running experiments
- Measure against established baselines
- Document expected outcomes and success criteria
- Run experiments regularly, not just after changes

### Result Analysis
- Compare against baseline metrics
- Identify root causes of degradation
- Implement improvements based on findings
- Retest to validate improvements

### Automation
- Integrate with CI/CD pipelines
- Schedule regular chaos experiments
- Automate report generation
- Track metrics over time

## Troubleshooting

### Common Issues

**Experiment not starting:**
```bash
# Check Chaos Mesh is deployed
kubectl get pods -n chaos-mesh

# Check CRDs are installed
kubectl get crds | grep chaos-mesh
```

**Experiments having no effect:**
```bash
# Verify selector matches target resources
kubectl get pods -l tier=frontend

# Check experiment status
kubectl describe networkchaos <experiment-name>
```

**Validation script errors:**
```bash
# Verify Python dependencies
pip install -r requirements.txt

# Check database connectivity
python3 -c "import psycopg2; print('PostgreSQL OK')"
```

### Getting Help
- Review documentation in `docs/chaos-engineering/`
- Check Chaos Mesh documentation: https://chaos-mesh.org/docs/
- Review experiment logs: `kubectl logs -n chaos-mesh <pod-name>`

## License

This chaos engineering implementation is part of the NEAM Platform and follows the same licensing terms.
