#!/usr/bin/env python3
"""
Analyze k6 benchmark results and generate comparison report
"""

import json
import sys
from pathlib import Path
from typing import Dict, Any

def load_summary(filepath: str) -> Dict[str, Any]:
    """Load k6 summary JSON file"""
    with open(filepath, 'r') as f:
        return json.load(f)

def extract_metrics(summary: Dict[str, Any]) -> Dict[str, Any]:
    """Extract key metrics from k6 summary"""
    metrics = summary.get('metrics', {})
    
    # HTTP metrics (k6 format: direct values, not nested in 'values')
    http_req_duration = metrics.get('http_req_duration', {})
    http_reqs = metrics.get('http_reqs', {})
    
    # Custom metrics  
    http_req_failed = metrics.get('http_req_failed', {})
    evaluation_duration = metrics.get('evaluation_duration', {})
    
    # Error rate from failed requests
    error_rate = http_req_failed.get('value', 0) if http_req_failed else 0
    
    # Note: k6 by default exports p(90) and p(95), not p(99)
    # Use p(90) if p(99) is not available
    p99 = http_req_duration.get('p(99)', None)
    if p99 is None:
        p99 = http_req_duration.get('p(90)', 0)  # Fallback to p90
    
    return {
        'requests_total': http_reqs.get('count', 0),
        'requests_per_sec': http_reqs.get('rate', 0),
        'latency_avg': http_req_duration.get('avg', 0),
        'latency_p50': http_req_duration.get('med', 0),
        'latency_p90': http_req_duration.get('p(90)', 0),
        'latency_p95': http_req_duration.get('p(95)', 0),
        'latency_p99': p99,
        'latency_max': http_req_duration.get('max', 0),
        'error_rate': error_rate,
        'evaluation_avg': evaluation_duration.get('avg', 0) if evaluation_duration else 0,
    }

def format_latency(ms: float) -> str:
    """Format latency in ms"""
    if ms == 0:
        return "N/A"
    return f"{ms:.2f}ms"

def format_rate(rate: float) -> str:
    """Format rate with percentage"""
    return f"{rate*100:.2f}%"

def calculate_advantage(val1: float, val2: float, lower_is_better: bool = True) -> str:
    """Calculate advantage factor - val1 is the 'bad' one, val2 is the 'good' one for lower_is_better"""
    if val1 == 0 or val2 == 0:
        return "N/A"
    
    if lower_is_better:
        # For latency: if camunda is slower (higher), dmn_go is faster
        factor = val1 / val2
        if factor > 1:
            return f"{factor:.2f}x faster"
        else:
            return f"{1/factor:.2f}x slower"
    else:
        # For throughput: higher is better
        factor = val1 / val2
        if factor > 1:
            return f"{factor:.2f}x higher"
        else:
            return f"{1/factor:.2f}x lower"

def main():
    if len(sys.argv) != 3:
        print("Usage: analyze_results.py <camunda_summary.json> <dmn_go_summary.json>")
        sys.exit(1)
    
    camunda_file = sys.argv[1]
    dmn_go_file = sys.argv[2]
    
    # Load summaries
    camunda_summary = load_summary(camunda_file)
    dmn_go_summary = load_summary(dmn_go_file)
    
    # Extract metrics
    camunda = extract_metrics(camunda_summary)
    dmn_go = extract_metrics(dmn_go_summary)
    
    # Generate report
    print("╔════════════════════════════════════════════════════════════════╗")
    print("║              PROFESSIONAL LOAD TEST COMPARISON                 ║")
    print("╠════════════════════════════════════════════════════════════════╣")
    print("")
    
    # THROUGHPUT - DMN Go first, then Camunda
    print("📊 THROUGHPUT & REQUESTS")
    print("─" * 68)
    print(f"{'Metric':<25} {'DMN Go':<16} {'Camunda 7':<16} {'Winner':<12}")
    print("─" * 68)
    
    # Throughput advantage calculation
    if dmn_go['requests_per_sec'] > camunda['requests_per_sec']:
        throughput_winner = f"DMN Go {dmn_go['requests_per_sec']/camunda['requests_per_sec']:.2f}x"
    else:
        throughput_winner = f"Camunda {camunda['requests_per_sec']/dmn_go['requests_per_sec']:.2f}x"
    
    print(f"{'Total Requests':<25} {dmn_go['requests_total']:<16} {camunda['requests_total']:<16}")
    print(f"{'Requests/sec':<25} {dmn_go['requests_per_sec']:<16.2f} {camunda['requests_per_sec']:<16.2f} {throughput_winner}")
    print("")
    
    # LATENCY - Lower is better, DMN Go should be lower
    print("⚡ LATENCY METRICS (Lower is Better)")
    print("─" * 68)
    print(f"{'Metric':<25} {'DMN Go':<16} {'Camunda 7':<16} {'Winner':<12}")
    print("─" * 68)
    
    # Calculate winners for each latency metric
    def latency_winner(dmn_val, cam_val):
        if dmn_val == 0 or cam_val == 0:
            return "N/A"
        if dmn_val < cam_val:
            return f"DMN Go {cam_val/dmn_val:.1f}x"
        else:
            return f"Camunda {dmn_val/cam_val:.1f}x"
    
    print(f"{'Average':<25} {format_latency(dmn_go['latency_avg']):<16} {format_latency(camunda['latency_avg']):<16} {latency_winner(dmn_go['latency_avg'], camunda['latency_avg'])}")
    print(f"{'Median (p50)':<25} {format_latency(dmn_go['latency_p50']):<16} {format_latency(camunda['latency_p50']):<16} {latency_winner(dmn_go['latency_p50'], camunda['latency_p50'])}")
    print(f"{'90th percentile (p90)':<25} {format_latency(dmn_go['latency_p90']):<16} {format_latency(camunda['latency_p90']):<16} {latency_winner(dmn_go['latency_p90'], camunda['latency_p90'])}")
    print(f"{'95th percentile (p95)':<25} {format_latency(dmn_go['latency_p95']):<16} {format_latency(camunda['latency_p95']):<16} {latency_winner(dmn_go['latency_p95'], camunda['latency_p95'])}")
    print(f"{'Max':<25} {format_latency(dmn_go['latency_max']):<16} {format_latency(camunda['latency_max']):<16} {latency_winner(dmn_go['latency_max'], camunda['latency_max'])}")
    print("")
    
    # RELIABILITY
    print("🎯 RELIABILITY")
    print("─" * 68)
    print(f"{'Metric':<25} {'DMN Go':<16} {'Camunda 7':<16}")
    print("─" * 68)
    print(f"{'Error Rate':<25} {format_rate(dmn_go['error_rate']):<16} {format_rate(camunda['error_rate']):<16}")
    print("")
    
    # SUMMARY
    print("📈 SUMMARY")
    print("─" * 68)
    
    # Count wins
    dmn_wins = 0
    camunda_wins = 0
    
    if dmn_go['requests_per_sec'] > camunda['requests_per_sec']:
        dmn_wins += 1
    else:
        camunda_wins += 1
        
    if dmn_go['latency_avg'] < camunda['latency_avg']:
        dmn_wins += 1
    else:
        camunda_wins += 1
        
    if dmn_go['latency_p95'] < camunda['latency_p95']:
        dmn_wins += 1
    else:
        camunda_wins += 1
    
    print(f"{'Throughput':<25} {'✅ DMN Go wins' if dmn_go['requests_per_sec'] > camunda['requests_per_sec'] else '✅ Camunda wins'}")
    print(f"{'Latency (avg)':<25} {'✅ DMN Go wins' if dmn_go['latency_avg'] < camunda['latency_avg'] else '✅ Camunda wins'}")
    print(f"{'Latency (p95)':<25} {'✅ DMN Go wins' if dmn_go['latency_p95'] < camunda['latency_p95'] else '✅ Camunda wins'}")
    print("")
    
    print("╚════════════════════════════════════════════════════════════════╝")
    print("")
    print("💡 Methodology: Professional load testing using k6")
    print("   - Sequential testing (no resource contention)")
    print("   - Realistic load profiles (ramp up/spike/sustained)")
    print("   - Identical resource constraints (2 CPU, 2GB RAM)")
    print("   - Same DMN definitions deployed to both systems")
    print("")

if __name__ == '__main__':
    main()
