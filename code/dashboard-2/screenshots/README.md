# Dashboard Screenshots

Screenshots of the NAIC Jobanalyzer Dashboard running with synthetic demo data.

## Cluster Selection

![Cluster Selection](01-cluster-selection.png)

Browse and select from available HPC clusters. Each cluster can have independent authentication via OIDC/PKCE.

## Cluster Overview

![Cluster Overview](02-cluster-overview.png)

Top-level cluster health: reporting nodes, total jobs, running/pending counts, and resource summaries.

![Cluster Overview - Resource Charts](02b-cluster-overview-scroll-1.png)

CPU and memory utilization timeseries across all nodes with interactive time range selection.

![Cluster Overview - GPU and Job Trends](02c-cluster-overview-scroll-2.png)

GPU utilization heatmap and job submission/completion trends over time.

![Cluster Overview - Queue and Disk I/O](02d-cluster-overview-scroll-3.png)

Queue wait time analysis by partition and disk I/O metrics.

![Cluster Overview - Bottom](02e-cluster-overview-scroll-4.png)

Additional cluster-wide metrics and resource distribution charts.

## Nodes

![Nodes](03-nodes.png)

Full node inventory with sortable/filterable AG Grid table showing hostname, state, CPU/memory/GPU specs, and current utilization.

## Partitions

![Partitions](04-partitions.png)

Split-panel partition browser with queue overview, node allocation, and GPU availability per partition.

## Jobs

![Jobs](05-jobs.png)

Live jobs table with status badges, resource allocation, elapsed time, and quick navigation to job details.

## Job Details

### Overview Tab

![Job Details - Overview](06a-job-details-overview.png)

Job summary with SAcct data, resource allocation breakdown, and efficiency metrics.

### Performance Metrics Tab

![Job Details - Performance Metrics](06b-job-details-performance.png)

CPU and memory efficiency gauges comparing allocated vs. used resources.

### Resource Timeline Tab

![Job Details - Resource Timeline](06c-job-details-timeline.png)

Time-series charts for CPU utilization, memory usage, memory utilization percentage, and process count over the job's lifetime.

### GPU Performance Tab

![Job Details - GPU Performance](06d-job-details-gpu.png)

Per-GPU utilization, memory usage, and performance metrics for jobs using GPU resources.

### Process Tree Tab

![Job Details - Process Tree](06e-job-details-process-tree.png)

Interactive process tree visualization showing the hierarchy from SLURM step daemon through to individual training workers and data loaders.

## Job Query

![Job Query - Form](07a-job-query-form.png)

Advanced job search form with filters for user, account, partition, job state, time range, and more.

![Job Query - Results](07b-job-query-results.png)

Query results displayed in a sortable AG Grid table with pagination.

## Benchmarks

![Benchmarks](08-benchmarks.png)

GPU benchmark comparisons (A100 vs H100) across ML workloads, filterable by precision and GPU count.
