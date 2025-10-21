- Score needs to be calculated on a case by case basis
- Meaning, a system with no GPU should not use GPU utilization in its calculation
- Probably need to use flags.
    - During collection, set flags when certain metrics are unavilable
    - During aggregation, flags will determine formula used to calculate Score

- GPU and CPU utilization needs to be tweaked based on # of GPUS and coresx
    - An 8 GPU system with 1 device at 100% is not as utilized as a 1 GPU system with its GPU at 100%
    