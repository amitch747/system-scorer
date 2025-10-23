# A Cluster Node System Metrics Prometheus Exporter


### Build Command
```
go build -o system-scraper && sudo mv system-scraper /usr/local/bin/ && sudo systemctl restart system-scraper && sudo systemctl restart prometheus
```

### system-scraper.service 
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