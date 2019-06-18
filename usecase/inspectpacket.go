package usecase

import (
	"arieoldman/arieoldman/krios/entity"
	"github.com/golang/glog"
)

// InspectPacket is a BooleanUseCase which inspects a packet and records results.
type InspectPacket struct {
	Conf entity.Config
}

// Execute causes the use case to execute and update the response handler.
func (uc InspectPacket) Execute(handler BooleanResponseHandler) {
	glog.Info("Just ran the InspectPacket use case.")
}
