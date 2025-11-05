# Score Exporter
**A node utilization-score exporter for HPC Slurm clusters**
## Scoring
### Weighted Score

$$
\huge utilization_{weighted} = 100 \times \left[1 - \prod_{i} (1 - w_i \cdot f_i)\right]
$$

### Weights
Weights (GPU nodes): $\large w_{GPU}=0.34, w_{CPU}=0.20, w_{mem}=0.10, w_{disk}=0.01, w_{net}=0.01, w_{user}=0.34$

Weights (CPU nodes): $\large w_{GPU}=0.00, w_{CPU}=0.54, w_{mem}=0.10, w_{disk}=0.01, w_{net}=0.01, w_{user}=0.34$



### Scaling Functions
Metrics used in scoring are transformed nonlinearly before aggregation:
- $\huge f_{GPU} = {gpu_{busy}}^{1.2}$
- $\huge f_{CPU} = {cpu_{busy}}^{1.2}$
- $\huge f_{Mem} = {mem_{usage}}^{1.5}$  
- $\huge f_{IO} = {io_{time}}^{1.2}$  
- $\huge f_{Net} = 1 - e^{-2 \cdot {net_{saturation}}}$
- $\huge f_{User} = users/capacity$




## Metrics
### Scoring

### GPU
- `gpu_busy_percent`
### CPU 
- `cpu_exec`
  - Percentage of scrape interval (15s default) of CPU time spent not in `idle` or `iowait` (0-100)
### Memory 
- `mem_usage`
  - Percentage of physical memory in use
### I/O 
- `io_time`
### Network 
- `net_saturation_percentage`
### Users
- `what_user_sessions_currently_active`
- `what_each_session_currently_active`