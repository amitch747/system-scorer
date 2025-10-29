package collector

import (
	"github.com/amitch747/system-scorer/utility"
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
			"syscore_net_saturation_percentage",
			"15s percentage of throughput over link capacity",
			[]string{"device"},
			nil,
		),
		netDropPercentage: prometheus.NewDesc(
			"syscore_net_drop_percentage",
			"15s percentage of packets dropped over total packets",
			[]string{"device"},
			nil,
		),
		netErrorPercentage: prometheus.NewDesc(
			"syscore_net_error_percentage",
			"15s percentage of packet errors over total packets",
			[]string{"device"},
			nil,
		),
	}
}

func (nc *networkCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(nc, ch)
}

type networkStats struct {
	bytesReceive, packetsReceive, errsReceive, dropReceive     uint64
	bytesTransmit, packetsTransmit, errsTransmit, dropTransmit uint64
}

type networkMetrics struct {
	satuationPercentage float64
	dropPercentage      float64
	errsPercentage      float64
}

var prevNetworkStats map[string]networkStats

func (nc *networkCollector) Collect(ch chan<- prometheus.Metric) {

	deviceNetStats, err := readNetworkStats()
	if err != nil {
		return
	}

	// Get device linkspeeds
	linkSpeeds := utility.GetLinkSpeeds()

	deviceMetrics := calcNetworkMetrics(deviceNetStats, linkSpeeds)

	prevNetworkStats = deviceNetStats

	for deviceName, device := range deviceMetrics {
		ch <- prometheus.MustNewConstMetric(
			nc.netSaturation,
			prometheus.GaugeValue,
			device.satuationPercentage,
			deviceName,
		)
		ch <- prometheus.MustNewConstMetric(
			nc.netDropPercentage,
			prometheus.GaugeValue,
			device.dropPercentage,
			deviceName,
		)
		ch <- prometheus.MustNewConstMetric(
			nc.netErrorPercentage,
			prometheus.GaugeValue,
			device.errsPercentage,
			deviceName,
		)
	}
}

func readNetworkStats() (map[string]networkStats, error) {
	deviceNetStats := map[string]networkStats{}

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

		deviceNetStats[deviceName] = networkStats{
			bytesReceive:   device.RxBytes,
			packetsReceive: device.RxPackets,
			errsReceive:    device.RxErrors,
			dropReceive:    device.RxDropped,

			bytesTransmit:   device.TxBytes,
			packetsTransmit: device.TxPackets,
			errsTransmit:    device.TxErrors,
			dropTransmit:    device.TxDropped,
		}
	}

	return deviceNetStats, nil
}

func calcNetworkMetrics(stats map[string]networkStats, linkSpeeds map[string]int64) map[string]networkMetrics {
	deviceMetrics := make(map[string]networkMetrics)

	if len(prevNetworkStats) == 0 {
		return deviceMetrics
	}
	// Have current. Need deltas
	for deviceName, netStats := range stats {
		// Filter out virtual? network devices
		if utility.NetDeviceFilter.MatchString(deviceName) {
			continue
		}
		// Get prev stats
		prevNetStats, ok := prevNetworkStats[deviceName]
		if !ok {
			continue
		}

		// Get link speed
		linkSpeed, ok := linkSpeeds[deviceName]
		if !ok || linkSpeed == 0 {
			continue
		}

		// Calculate saturation
		deltaRxBytes := netStats.bytesReceive - prevNetStats.bytesReceive
		deltaTxBytes := netStats.bytesTransmit - prevNetStats.bytesTransmit
		totalBytes := deltaRxBytes + deltaTxBytes
		throughputBps := float64(totalBytes) / utility.ScrapeInterval
		saturationPercentage := (float64(throughputBps) / float64(linkSpeed)) * 100.0

		deltaRxPackets := netStats.packetsReceive - prevNetStats.packetsReceive
		deltaTxPackets := netStats.packetsTransmit - prevNetStats.packetsTransmit
		totalPackets := float64(deltaRxPackets + deltaTxPackets)

		deltaRxDrop := float64(netStats.dropReceive - prevNetStats.dropReceive)
		deltaTxDrop := float64(netStats.dropTransmit - prevNetStats.dropTransmit)

		deltaRxError := float64(netStats.errsReceive - prevNetStats.errsReceive)
		deltaTxError := float64(netStats.errsTransmit - prevNetStats.errsTransmit)

		var dropPercentage, errPercentage float64

		if totalPackets > 0 {
			dropPercentage = ((deltaRxDrop + deltaTxDrop) / totalPackets) * 100.0
			errPercentage = ((deltaRxError + deltaTxError) / totalPackets) * 100.0
		} else {
			dropPercentage = 0.0
			errPercentage = 0.0
		}

		deviceMetrics[deviceName] = networkMetrics{
			satuationPercentage: saturationPercentage,
			dropPercentage:      dropPercentage,
			errsPercentage:      errPercentage,
		}
		// Take max?

	}
	return deviceMetrics
}

//type netDevStats map[string]map[string]uint64
// /proc/net/dev
// 15s delta just like cpu.go

// Bandwidth
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
