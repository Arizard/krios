package infrastructure

import (
	"arieoldman/arieoldman/krios/entity"
	"fmt"
	"os"
)

// FileReportRepository is a ReportRepository which saves to a file (csv)
type FileReportRepository struct {
	Stream *os.File
}

// Add converts a report to csv formatting and appends it to the stream.
func (repo *FileReportRepository) Add(report entity.Report) {
	for _, intel := range report.Intels {
		fmt.Fprintf(
			repo.Stream,
			"%d,%x,%x,%q,%q,%d,%d,%d,%d,%d,,,,\n",
			intel.Timestamp,
			intel.SrcMAC,
			intel.DstMAC,
			intel.SrcIP,
			intel.DstIP,
			intel.SrcTCP,
			intel.DstTCP,
			intel.SrcUDP,
			intel.DstUDP,
			intel.Size,
		)
	}
}
