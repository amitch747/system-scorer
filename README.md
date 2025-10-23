# A Cluster Node System Metrics Prometheus Exporter


GPU node
Score=100×(0.45*gpu​+0.20*cpu​+0.15*mem​+0.10*disk​+0.10*net​)

CPU node
Score=100×(0.50*cpu​+0.25*mem​+0.15*disk​+0.10*net​)


cpu = (cpu_util_exec/100)^1.2
gpu = (0.7*gpu_busy + 0.3*gpu_mem_used)^1.2
mem = (mem_used)^1.5
disk = (disk_busy)^1.2
net = (net_util)^1.2