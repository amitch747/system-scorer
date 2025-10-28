package collector

type networkCollector struct {
}

// /proc/net/dev
// 15s delta just like cpu.go

// Bandwidth saturation
/*
delta_rx := curr.rx_bytes - prev.rx_bytes
delta_tx := curr.tx_bytes - prev.tx_bytes
total_bytes := delta_rx + delta_tx
throughput_Bps := total_bytes / scrapeInterval

// /sys/class/net/<iface>/speed
link_capacity_Bps := speed_Mbps * 125000

saturation = throughput_Bps / link_capacity_Bps
*/

// Drop Ratio (congestion)
/*
delta_rx_drop := curr.rx_drop - prev.rx_drop
delta_tx_drop := curr.tx_drop - prev.tx_drop
delta_rx_pkts := curr.rx_packets - prev.rx_packets
delta_tx_pkts := curr.tx_packets - prev.tx_packets

drop_ratio = (delta_rx_drop + delta_tx_drop) / (delta_rx_pkts + delta_tx_pkts)

*/

// Error Ratio (health)
// Same math as above

// Final pressure
/*
scaledUtil  := 1 - math.Exp(-4 * utilization)
scaledDrop  := math.Min(1, drop_ratio * 10)
scaledError := math.Min(1, error_ratio * 10)
netPressure := 0.8*scaledUtil + 0.15*scaledDrop + 0.05*scaledError
*/
