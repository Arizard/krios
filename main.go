package main

import (
	// "arieoldman/arieoldman/krios/common"
	"arieoldman/arieoldman/krios/controller"
	"arieoldman/arieoldman/krios/entity"
	"arieoldman/arieoldman/krios/infrastructure"
	"flag"
	"github.com/golang/glog"
	"time"
	"fmt"
)

func main() {
	flag.Parse()
	var cp entity.ControlPlane

	conf := entity.Config{
	}

	ctrl := controller.Session{
		Conf: conf,
	}
	ctrl.Initialise()

	cp = &infrastructure.OpenFlow13ControlPlane{
		//Session: ctrl,
	}

	cp.Setup()

	cp.SetupLayer2Switching()

	go cp.Start(6633)

	go cliLoop(100 * time.Millisecond)

	for {

	}

	glog.Info("Finished.")

	glog.Flush()
}

func cliLoop(delay time.Duration){
	for {
		for _,r := range `–\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
