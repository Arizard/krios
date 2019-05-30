package main

import (
	"arieoldman/arieoldman/krios/controller"
	"arieoldman/arieoldman/krios/entity"
	"arieoldman/arieoldman/krios/infrastructure"
	"flag"
	"github.com/golang/glog"
	"time"
)

func main() {
	flag.Parse()
	var cp entity.ControlPlane

	conf := entity.Config{}

	ctrl := controller.Session{
		Conf: conf,
	}
	ctrl.Initialise()

	cp = &infrastructure.OpenFlow13ControlPlane{
		//Session: ctrl,
	}

	cp.Setup()

	cp.SetupLayer2Switching()

	cp.Start(6633)

	for {
		time.Sleep(1 * time.Second)
	}

	glog.Info("Finished.")

	glog.Flush()
}
