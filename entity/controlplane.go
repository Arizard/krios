package entity

// ControlPlane defines the contract for an SDN control - data plane link.
type ControlPlane interface {
	Start(port uint16)
	Stop()
	Setup()
	SetupLayer2Switching()
}
