package infrastructure

import (
	"arieoldman/arieoldman/krios/entity"
	"fmt"
	"io"
	"github.com/golang/glog"
)

// FileReportRepository is a ReportRepository which saves to a file (csv)
type FileReportRepository struct {
	Stream io.Writer
}

// Add converts a report to csv formatting and appends it to the stream.
func (repo *FileReportRepository) Add(report entity.Report) {
	for _, intel := range report.Intels {
		n, err := fmt.Fprintf(
			repo.Stream,
			"%d,%x,%x,%q,%q,%d,%d,%d,%d,%d,%s,\n",
			intel.Timestamp,
			[]byte(intel.SrcMAC),
			[]byte(intel.DstMAC),
			intel.SrcIP,
			intel.DstIP,
			intel.SrcTCP,
			intel.DstTCP,
			intel.SrcUDP,
			intel.DstUDP,
			intel.Size,
			intel.Detail,
		)

		glog.Infof("Wrote intel: %d %v", n, err)
	}
}
