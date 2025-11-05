# Score Exporter
**A node utilization-score exporter for HPC Slurm clusters**
## Scoring

### Scaling Functions

Metrics used in scoring are transformed nonlinearly before aggregation:
- **GPU:** $\huge g = {gpu_{busy}}^{1.2}$
- **CPU:** $\huge c = {cpu_{busy}}^{1.2}$
- **Memory:** $\huge m = {mem_{usage}}^{1.5}$  
- **Disk I/O:** $\huge i = {io_{time}}^{1.2}$  
- **Network:** $\huge n = 1 - e^{-2 \cdot {net_{saturation}}}$
- **Users:** $\huge user\#$

### Weighted Score

$$
\huge utilization_{weighted} = 100 \times \left[1 - \prod_{i} (1 - w_i \cdot f_i)\right]
$$

Weights (GPU nodes): $\large w_{GPU}=0.34, w_{CPU}=0.20, w_{mem}=0.10, w_{disk}=0.01, w_{net}=0.01, w_{user}=0.34$
Weights (CPU nodes): $\large w_{GPU}=0.00, w_{CPU}=0.54, w_{mem}=0.10, w_{disk}=0.01, w_{net}=0.01, w_{user}=0.34$


## Metrics (WIP)
### GPU
- `gpu_busy_percent`
### CPU 
- `cpu_exec`
  - Percentage of scrape interval (15s default) of CPU time spent not in `idle` or `iowait` (0-100)
### Users (WIP)
- `what_user_sessions_currently_active`
- `what_each_session_currently_active`
### Memory 
- `mem_usage`
  - Percentage of physical memory in use
### I/O 
- `io_time`
### Network 
- `net_saturation_percentage`