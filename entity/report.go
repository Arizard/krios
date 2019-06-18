package entity

import (
	"net"
	"time"
)

// Intel is an entity which describes the intelligence data for a packet.
type Intel struct {
	SrcMAC    net.HardwareAddr
	DstMAC    net.HardwareAddr
	SrcIP     net.IP
	DstIP     net.IP
	SrcTCP    uint16
	DstTCP    uint16
	SrcUDP    uint16
	DstUDP    uint16
	Size      uint16
	Timestamp int64 // nanoseconds since unix epoch
}

// NewReport creates a new report instance.
func NewReport() Report {
	return Report{
		Intels: []Intel{},
	}
}

// Report is an aggregate of ordered Intel entities.
type Report struct {
	Intels []Intel
}

// AddIntel adds a new Intel instance to the report.
func (rep *Report) AddIntel(
	srcMAC net.HardwareAddr,
	dstMAC net.HardwareAddr,
	srcIP net.IP,
	dstIP net.IP,
	srcTCP uint16,
	dstTCP uint16,
	size uint16,
) {
	rep.Intels = append(rep.Intels, Intel{
		SrcMAC: srcMAC,
		DstMAC: dstMAC,
		SrcIP: srcIP,
		DstIP: dstIP,
		SrcTCP: srcTCP,
		DstTCP: dstTCP,
		Size: size,
		Timestamp: time.Now().UnixNano(),
	})
}

// ReportRepository defines the methods for storing reports.
type ReportRepository interface {
	Add(report Report)
}
