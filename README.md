# Node Utilization Score Exporter for Slurm Clusters
Description work in progress ;)
## Scoring

### Scaling Functions

Metrics used in scoring are transformed nonlinearly before aggregation:
- **GPU:** $\huge g = {gpu_{busy}}^{1.2}$
- **CPU:** $\huge c = {cpu_{busy}}^{1.2}$
- **Memory:** $\huge m = {mem_{usage}}^{1.5}$  
- **Disk I/O:** $\huge i = {io_{time}}^{1.2}$  
- **Network:** $\huge n = 1 - e^{-2 \cdot {net_{saturation}}}$
- **Users:** $\huge \# users$

### Weighted Score

$$
\huge utilization_{weighted} = 100 \times \left[1 - \prod_{i} (1 - w_i \cdot f_i)\right]
$$

Weights (GPU nodes): $\large w_{GPU}=0.34, w_{CPU}=0.20, w_{mem}=0.10, w_{disk}=0.01, w_{net}=0.01, w_{user}=0.34$

Weights (CPU nodes): $\large w_{GPU}=0.00, w_{CPU}=0.54, w_{mem}=0.10, w_{disk}=0.01, w_{net}=0.01, w_{user}=0.34$

### Bottleneck Score

$$
\huge utilization_{bottleneck} = 100 \times \max(c, m, d, g, n, u)
$$

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

## Setup (WIP)
### Create systemd service
`sudo nano /etc/systemd/system/system-scorer.service`
```
[Unit]
Description=System Scorer Prometheus Exporter
After=network.target

[Service]
ExecStart=/usr/local/bin/system-scorer
WorkingDirectory=/usr/local/bin
# Optional
EnvironmentFile=-/etc/system-scorer/env.conf
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### Build binary and start service

`go build -o system-scorer ./cmd && sudo mv system-scorer /usr/local/bin/`

`sudo systemctl restart system-scorer`


## Viewing  
### Local
`curl http://localhost:8081/metrics`

### Prometheus
- Update `/etc/prometheus/prometheus.yml`
```
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'system_scraper'
    static_configs:
      - targets: ['localhost:8081']
```

```
sudo systemctl daemon-reload
sudo systemctl restart prometheus
sudo systemctl status prometheus
```
http://localhost:9090/classic/targets


### Grafana

```
sudo apt update
sudo apt install grafana

sudo systemctl enable grafana-server
sudo systemctl start grafana-server
sudo systemctl status grafana-server
```
- Default credentials will be admin/admin
- Go to Settings -> Data sources -> Add data source -> Prometheus
- URL: http://localhost:9090
- Click save and test

- Grafana will be at http://localhost:3000
