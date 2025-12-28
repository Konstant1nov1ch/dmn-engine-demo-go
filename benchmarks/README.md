# Professional Load Testing with k6

## Overview

Professional benchmark suite using industry-standard tools:
- **k6**: Modern load testing tool (by Grafana Labs)
- **Python**: Data analysis and reporting
- **Docker**: Isolated environment management

## Prerequisites

### Install k6

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```bash
choco install k6
```

**Or use Docker:**
```bash
docker pull grafana/k6:latest
```

## Quick Start

```bash
# Run benchmark with 50 decisions
./benchmarks/run_benchmark.sh 50

# Or more decisions for stress test
./benchmarks/run_benchmark.sh 200
```

## Fair Comparison: Both in Docker

✅ **DMN Engine Go**: Docker container (2GB RAM, 2 CPU limit)
✅ **Camunda 7**: Docker container (2GB RAM, 2 CPU limit)
✅ **PostgreSQL**: Separate containers for each
✅ **Equal conditions**: Same resource limits, same network setup

This ensures an **apples-to-apples** comparison.

## What Gets Tested

### Load Profile (5 minutes total)

1. **Ramp Up (30s)**: 0 → 10 users
2. **Sustained (1m)**: 10 → 50 users  
3. **Peak (2m)**: 50 users (steady state)
4. **Spike (30s)**: 50 → 100 users
5. **Sustained Peak (1m)**: 100 users
6. **Ramp Down (30s)**: 100 → 0 users

### Metrics Collected

#### HTTP Metrics
- Requests per second (throughput)
- Total requests
- HTTP response times

#### Latency Metrics
- Average latency
- Median (p50)
- 95th percentile (p95)
- 99th percentile (p99)
- Max latency

#### Reliability
- Error rate
- Success rate
- Response validation

#### System Metrics
- Memory usage (before/after)
- CPU usage (via Docker stats)

## Results

Results are saved to `benchmarks/results/` with timestamp:
- `camunda_YYYYMMDD_HHMMSS.json` - Full k6 metrics
- `dmn_go_YYYYMMDD_HHMMSS.json` - Full k6 metrics
- `camunda_summary_YYYYMMDD_HHMMSS.json` - Summary statistics
- `dmn_go_summary_YYYYMMDD_HHMMSS.json` - Summary statistics
- `comparison_YYYYMMDD_HHMMSS.txt` - Side-by-side comparison

### Example Output

```
╔════════════════════════════════════════════════════════════════╗
║              PROFESSIONAL LOAD TEST COMPARISON                 ║
╠════════════════════════════════════════════════════════════════╣

📊 THROUGHPUT & REQUESTS
────────────────────────────────────────────────────────────────
Metric                    DMN Go          Camunda 7       Advantage 
────────────────────────────────────────────────────────────────
Total Requests            15234           15189          
Requests/sec              254.56          253.15          1.01x higher

⚡ LATENCY METRICS (Lower is Better)
────────────────────────────────────────────────────────────────
Metric                    DMN Go          Camunda 7       Advantage 
────────────────────────────────────────────────────────────────
Average                   12.45ms         145.67ms        11.70x faster
Median (p50)              10.23ms         134.12ms        13.11x faster
95th percentile (p95)     25.89ms         298.45ms        11.53x faster
99th percentile (p99)     45.67ms         456.78ms        10.00x faster
Max                       123.45ms        789.12ms        

🎯 RELIABILITY
────────────────────────────────────────────────────────────────
Metric                    DMN Go          Camunda 7       
────────────────────────────────────────────────────────────────
Error Rate                0.12%           0.45%           

╚════════════════════════════════════════════════════════════════╝

💡 Methodology: Professional load testing using k6
   - Sequential testing (no resource contention)
   - Realistic load profiles (ramp up/spike/sustained)
   - Statistical significance (p50, p95, p99)
```

## Custom Load Profiles

Edit `benchmarks/k6/*.js` to customize:

```javascript
export const options = {
  stages: [
    { duration: '1m', target: 100 },  // Your custom profile
    { duration: '5m', target: 100 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],  // Your SLA
    errors: ['rate<0.01'],
  },
};
```

## CI/CD Integration

```yaml
# .github/workflows/benchmark.yml
- name: Run Benchmarks
  run: |
    ./benchmarks/run_benchmark.sh 100
    
- name: Upload Results
  uses: actions/upload-artifact@v3
  with:
    name: benchmark-results
    path: benchmarks/results/
```

## Why k6?

✅ **Industry Standard**: Used by Grafana, Microsoft, Apple  
✅ **Modern**: Written in Go, performant  
✅ **Flexible**: JavaScript configuration  
✅ **Professional Metrics**: p95, p99 out of the box  
✅ **Integrations**: Grafana, InfluxDB, Prometheus  
✅ **Cloud**: Cloud service available  

## Advanced Features

### Export to Grafana

```bash
k6 run \
  --out influxdb=http://localhost:8086/k6 \
  benchmarks/k6/dmn_go_load_test.js
```

### Cloud Monitoring

```bash
k6 cloud benchmarks/k6/dmn_go_load_test.js
```

### Thresholds & SLAs

Define SLAs in test scripts:
```javascript
thresholds: {
  'http_req_duration': ['p(95)<500', 'p(99)<1000'],
  'errors': ['rate<0.05'],
  'http_req_duration{status:200}': ['max<2000'],
}
```

## Documentation

- [k6 Documentation](https://k6.io/docs/)
- [k6 Examples](https://k6.io/docs/examples/)
- [Best Practices](https://k6.io/docs/testing-guides/load-testing-best-practices/)

