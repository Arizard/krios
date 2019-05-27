package infrastructure

// ControlPlane defines the contract for an SDN control - data plane link.
type ControlPlane interface {
	Start(port uint16)
	Stop()
}

// OpenFlowControlPlane defines the contract for an OpenFlow control plane.
type OpenFlowControlPlane interface {
	ControlPlane
}