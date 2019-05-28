package main

import (
	"flag"
	"github.com/golang/glog"
	"arieoldman/arieoldman/krios/controller"
	"arieoldman/arieoldman/krios/common"
	"arieoldman/arieoldman/krios/entity"
	"arieoldman/arieoldman/krios/infrastructure"
)


func main() {
	flag.Parse()
	var cp entity.ControlPlane

	conf := entity.Config{
		DPIDs: []common.EthAddr{
			common.EthAddr{ 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			common.EthAddr{ 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
			common.EthAddr{ 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
		},
	}

	ctrl := controller.Session{
		Conf: conf,
	}
	ctrl.Initialise()

	cp = &infrastructure.OpenFlow13ControlPlane{}

	cp.Setup()

	cp.SetupLayer2Switching()

	cp.Start(6633)

	glog.Info("Finished.")

	glog.Flush()
}