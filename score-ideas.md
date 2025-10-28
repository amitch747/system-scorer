# Utilization
- Score needs to be calculated on a case by case basis
- Meaning, a system with no GPU should not use GPU utilization in its calculation
- Probably need to use flags.
    - During collection, set flags when certain metrics are unavilable
    - During aggregation, flags will determine formula used to calculate Score

- GPU and CPU utilization needs to be tweaked based on # of GPUS and cores
    - An 8 GPU system with 1 device at 100% is not as utilized as a 1 GPU system with its GPU at 100%
    



- cpu = (cpu_util_exec/100)^1.2
- gpu = (0.7*gpu_busy + 0.3*gpu_mem_used)^1.2
- mem = (mem_used)^1.5
- disk = (disk_busy)^1.2
- net = (net_util)^1.2



- Slurm stuff needs to be factored in at the end. Reservations and jobs (even if they are not actually using resources) are technically utilizing the system
    - Someone may reserve a node for 24hrs and do NOTHING, but as far as everyone else is concerned, that node is not usable

