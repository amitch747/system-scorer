# A Cluster Node System Metrics Prometheus Exporter


## Setup
### Create systemd service
`sudo nano /etc/systemd/system/system-scraper.service`
```
[Unit]
Description=System Scraper Exporter
After=network.target

[Service]
ExecStart=/usr/local/bin/system-scraper
WorkingDirectory=/usr/local/bin
# Optional
EnvironmentFile=-/etc/system-scraper/env.conf
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
User=root

[Install]
WantedBy=multi-user.target
```

### Build binary and start service

`go build -o cmd/system-scraper && sudo mv cmd/system-scraper /usr/local/bin/`

`sudo systemctl restart system-scraper`


## Viewing Metrics
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
