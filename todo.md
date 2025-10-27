### Features
- Fix I/O
- Network
- Normalize all values
- Update readme with proper build instructions, and math

- Figure out why I can't get gpu stats
- Dynamically check if node has GPU or not
- GPU

### Bugs
- Fix CPU name and numbers
- Change sessions back to using `who` instead of `w` since apparently its being disabled. Thus, refactor whole collector.
- Remove other exporters from /etc/prometheus/prometheus.yml
- Swap collect and describe recievers to use pointers