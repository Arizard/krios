package controller

import (
	"arieoldman/arieoldman/krios/entity"
	"github.com/golang/glog"
)

// SessionManager defines the behaviour of a controller session.
type SessionManager interface {
	Initialise()
}

// Session is the concrete controller session for the application.
// Session is a SessionManager
type Session struct {
	Conf entity.Config
	ControlPlane entity.ControlPlane
	ReportRepository entity.ReportRepository
}

// Initialise begins the session and sets up necessary things.
func (s Session) Initialise() {
	glog.Infof("Session initialised.\n")
	s.ControlPlane.Setup()
	if s.Conf.L2Switching {
		s.ControlPlane.SetupLayer2Switching()
	}
	if s.Conf.DPIEnabled {
		s.ControlPlane.SetupDeepPacketInspection(s.ReportRepository)
	}
	defer s.ControlPlane.Start(6633)

}