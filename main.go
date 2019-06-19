package main

import (
	"arieoldman/arieoldman/krios/controller"
	"arieoldman/arieoldman/krios/entity"
	"arieoldman/arieoldman/krios/infrastructure"
	"flag"
	"github.com/golang/glog"
	"os"
	"bufio"
)

var flag_switch = flag.Bool("switch", false, "Enables switch mode.")
var flag_dpi = flag.Bool("dpi", false, "Enables deep packet inspector.")

func main() {
	defer glog.Flush()
	flag.Parse()
	
	var conf entity.Config
	var cp entity.ControlPlane
	var repRepo entity.ReportRepository
	var ctrl controller.SessionManager

	conf = entity.Config{
		L2Switching: *flag_switch,
		DPIEnabled: *flag_dpi,
	}
	
	cp = &infrastructure.OpenFlow13ControlPlane{}

	repFile, err := os.Create("report.log")
	defer repFile.Close()

	repWriter := bufio.NewWriter(repFile)
	defer repWriter.Flush()

	if err != nil {
		panic(err)
	}

	repRepo = &infrastructure.FileReportRepository{
		Stream: repFile,
	}
	
	ctrl = controller.Session{
		Conf: conf,
		ControlPlane: cp,
		ReportRepository: repRepo,
	}
	ctrl.Initialise()

	glog.Info("Finished.")

	
}
