package controller

import (
	"arieoldman/arieoldman/krios/common"
	"arieoldman/arieoldman/krios/entity"
	"arieoldman/arieoldman/krios/usecase"
)

// SessionManager defines the behaviour of a controller session.
type SessionManager interface {
	Initialise()
	CanHandshake(dpid common.EthAddr) bool
}

// Session is the concrete controller session for the application.
// Session is a SessionManager
type Session struct {
	Conf entity.Config
}

// Initialise begins the session and sets up necessary things.
func (s *Session) Initialise() {

}

type canHandshakeResponseHandler struct {
	response bool
}

func (handler *canHandshakeResponseHandler) Handle(b bool) {
	handler.response = b
}

func (handler *canHandshakeResponseHandler) Response() bool {
	return handler.response
}

// CanHandshake returns true if the dpid is allowed to handshake.
func (s *Session) CanHandshake(dpid common.EthAddr) bool {
	var handler usecase.BooleanResponseHandler //interfaces
	var uc usecase.BooleanUseCase

	handler = &canHandshakeResponseHandler{}

	// "Datapath ... is a BooleanUseCase"
	uc = usecase.DatapathCanHandshake{
		Conf: s.Conf,
		DPID: dpid,
	}

	uc.Execute(handler)

	return handler.Response()
}
