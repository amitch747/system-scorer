package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

type networkCollector struct {
	netSaturation      *prometheus.Desc
	netDropPercentage  *prometheus.Desc
	netErrorPercentage *prometheus.Desc
}

func NewNetworkCollector() *networkCollector {
	return &networkCollector{
		netSaturation: prometheus.NewDesc(
			"syscore_net_throughput_percentage",
			"15s percentage of throughput over link capacity",
			nil,
			nil,
		),
		netDropPercentage: prometheus.NewDesc(
			"syscore_net_drop_percentage",
			"15s percentage of packets dropped over total packets",
			nil,
			nil,
		),
		netErrorPercentage: prometheus.NewDesc(
			"syscore_net_error_percentage",
			"15s percentage of packet errors over total packets",
			nil,
			nil,
		),
	}
}

func (nc *networkCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(nc, ch)
}

type networkStats struct {
	bytesReceive, packetsReceive, errsReceive, dropRecieve     uint64
	bytesTransmit, packetsTransmit, errsTransmit, dropTransmit uint64
}

var prevNetworkStats map[string]networkStats

func (nc *networkCollector) Collect(ch chan<- prometheus.Metric) {
	networkStats, err := readNetworkStats()
	if err != nil {
		return
	}
	println(networkStats)

	ch <- prometheus.MustNewConstMetric(
		nc.netSaturation,
		prometheus.GaugeValue,
		float64(2.2),
	)
	ch <- prometheus.MustNewConstMetric(
		nc.netDropPercentage,
		prometheus.GaugeValue,
		float64(2.2),
	)
	ch <- prometheus.MustNewConstMetric(
		nc.netErrorPercentage,
		prometheus.GaugeValue,
		float64(2.2),
	)
}

func readNetworkStats() (map[string]networkStats, error) {
	deviceStats := map[string]networkStats{}

	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, err
	}
	// https://pkg.go.dev/github.com/prometheus/procfs#FS.NetDev
	procNetDev, err := fs.NetDev()
	if err != nil {
		return nil, err
	}

	for _, device := range procNetDev {
		deviceName := device.Name

		deviceStats[deviceName] = networkStats{
			bytesReceive:   device.RxBytes,
			packetsReceive: device.RxPackets,
			errsReceive:    device.RxErrors,
			dropRecieve:    device.RxDropped,

			bytesTransmit:   device.TxBytes,
			packetsTransmit: device.TxPackets,
			errsTransmit:    device.TxErrors,
			dropTransmit:    device.TxDropped,
		}
	}

	return deviceStats, nil
}

//type netDevStats map[string]map[string]uint64
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
