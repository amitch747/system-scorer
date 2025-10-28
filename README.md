## Scores (work in progress)
### System Utilization
- GPU node
    - Score=100×(0.45*gpu​+0.20*cpu​+0.15*mem​+0.10*disk​+0.10*net​)
    - Include user sessions?

- CPU node
    - Score=100×(0.50*cpu​+0.25*mem​+0.15*disk​+0.10*net​)
### System Health

## Metrics
### GPU (WIP)
### CPU 
- `cpu_exec_percentage`
  - CPU time spent not in `idle` or `iowait`
### Memory 
- `mem_usage`
  - Percentage of physical memory in use
- `mem_commit`
  - Percentage of committed virtual memory over commit limit
- `mem_swap`
  - Percentage of swap space in use
- `mem_pressure`
  - Weighted memory pressure index (usage + swap + commit)
### I/O (WIP)
- `io_time`
- `io_pressure`
### Network (WIP)
### Users (WIP)

- `what_user_sessions_currently_active`
- `what_each_session_currently_active`




- `users_total`
  - Number of logged in users.
- `user_session_count`
  - Number of active sessions per user.

## Setup 
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
- Update config
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
http://localhost:3000
- Default credentials will be admin/admin
- Go to Settings -> Data sources -> Add data source -> Prometheus
- URL: http://localhost:9090
- Click save and test
