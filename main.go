package main

import (
	"arieoldman/arieoldman/krios/controller"
	"arieoldman/arieoldman/krios/entity"
	"arieoldman/arieoldman/krios/infrastructure"
	"flag"
	"github.com/golang/glog"
	"os"
)

var flag_switch = flag.Bool("switch", false, "Enables switch mode.")

func main() {
	flag.Parse()
	
	var conf entity.Config
	var cp entity.ControlPlane
	var repRepo entity.ReportRepository
	var ctrl controller.SessionManager

	conf = entity.Config{
		L2Switching: *flag_switch,
		DPIEnabled: true,
	}
	
	cp = &infrastructure.OpenFlow13ControlPlane{}

	repRepo = &infrastructure.FileReportRepository{
		Stream: os.Stdout,
	}
	
	ctrl = controller.Session{
		Conf: conf,
		ControlPlane: cp,
		ReportRepository: repRepo,
	}
	ctrl.Initialise()

	glog.Info("Finished.")

	glog.Flush()
}
