package infrastructure

import (
	// "arieoldman/arieoldman/krios/common"
	"arieoldman/arieoldman/krios/controller"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	of "github.com/netrack/openflow"
	"github.com/netrack/openflow/ofp"
	"github.com/netrack/openflow/ofputil"
)

type openFlowEventHook struct {
}

// OpenFlow13ControlPlane is an OpenFlow 1.3 control plane.
type OpenFlow13ControlPlane struct {
	ctrlSession controller.SessionManager
	mux         *of.ServeMux

	// TODO: Create a hooks system which allows applications to
	// "hook" functions into an openflow event. This way multiple functions
	// can be run by a single handler on a single mux.
	// Needs a mapping of of.TypeMatcher to a slice of functions (possible?)
	// When an event occurs, all of the functions mappet to the of.TypeMatcher
	// are executed in order of when they were added.

	customHandlers map[of.TypeMatcher]([]of.HandlerFunc)
}

func (cp *OpenFlow13ControlPlane) customHandleFunc(tm of.TypeMatcher, h of.HandlerFunc) {
	if cp.customHandlers[tm] == nil {
		cp.customHandlers[tm] = []of.HandlerFunc{}
	}
	cp.customHandlers[tm] = append(cp.customHandlers[tm], h)
}

const (
	echoRequestEvent   of.TypeMatcher = of.TypeMatcher(of.TypeEchoRequest)
	featuresReplyEvent of.TypeMatcher = of.TypeMatcher(of.TypeFeaturesReply)
	helloEvent         of.TypeMatcher = of.TypeMatcher(of.TypeHello)
	errorEvent         of.TypeMatcher = of.TypeMatcher(of.TypeError)
	packetInEvent      of.TypeMatcher = of.TypeMatcher(of.TypePacketIn)

	ctrlTable ofp.Table = ofp.Table(0)
	fwdTable  ofp.Table = ofp.Table(1)
)

var (
	gotoForwardingTable = &ofp.InstructionGotoTable{
		fwdTable,
	}

	sendController = &ofp.InstructionApplyActions{
		ofp.Actions{
			&ofp.ActionOutput{ofp.PortController, ofp.ContentLenMax},
		},
	}

	floodPorts = &ofp.InstructionApplyActions{
		ofp.Actions{
			&ofp.ActionOutput{ofp.PortFlood, 0},
		},
	}

	dropPacket = &ofp.InstructionApplyActions{
		ofp.Actions{
			//Empty actions means we drop the packet.
		},
	}

	matchEverything = ofp.XM{
		Class: ofp.XMClassOpenflowBasic,
		Type:  ofp.XMTypeEthDst,
		Value: ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Mask:  ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	matchEthType0800 = ofp.XM{
		Class: ofp.XMClassOpenflowBasic,
		Type:  ofp.XMTypeEthType,
		Value: ofp.XMValue{0x08, 0x00},
	}

	matchIPProto6 = ofp.XM{
		Class: ofp.XMClassOpenflowBasic,
		Type:  ofp.XMTypeIPProto,
		Value: ofp.XMValue{0x06},
	}
	matchHTTP = ofp.XM{
		Class: ofp.XMClassOpenflowBasic,
		Type:  ofp.XMTypeTCPDst,
		Value: ofp.XMValue{0x00, 0x50}, // 80 in base 16
	}
)

// Setup initialises all the stuff we need.
func (cp *OpenFlow13ControlPlane) Setup() {
	cp.mux = of.NewServeMux()
	cp.customHandlers = make(map[of.TypeMatcher]([]of.HandlerFunc))

	cp.mux.HandleFunc(errorEvent, func(rw of.ResponseWriter, r *of.Request) {
		for _, h := range cp.customHandlers[errorEvent] {
			h(rw, r)
		}
	})

	cp.mux.HandleFunc(featuresReplyEvent, func(rw of.ResponseWriter, r *of.Request) {
		for _, h := range cp.customHandlers[featuresReplyEvent] {
			h(rw, r)
		}
	})

	cp.mux.HandleFunc(helloEvent, func(rw of.ResponseWriter, r *of.Request) {
		for _, h := range cp.customHandlers[helloEvent] {
			h(rw, r)
		}
	})

	cp.mux.HandleFunc(echoRequestEvent, func(rw of.ResponseWriter, r *of.Request) {
		for _, h := range cp.customHandlers[echoRequestEvent] {
			h(rw, r)
		}
	})

	cp.mux.HandleFunc(packetInEvent, func(rw of.ResponseWriter, r *of.Request) {
		for _, h := range cp.customHandlers[packetInEvent] {
			h(rw, r)
		}
	})
}

// Start will start the control plane listener
func (cp *OpenFlow13ControlPlane) Start(port uint16) {

	cp.customHandleFunc(errorEvent, func(rw of.ResponseWriter, r *of.Request) {
		var packet ofp.Error
		packet.ReadFrom(r.Body)

		glog.Errorln("Error:", packet.Error())
	})

	cp.customHandleFunc(featuresReplyEvent, func(rw of.ResponseWriter, r *of.Request) {
		var featuresReply ofp.SwitchFeatures
		featuresReply.ReadFrom(r.Body)

		glog.Infof("Features Reply from %s: DatapathID %x, %v\n",
			r.Addr, featuresReply.DatapathID, featuresReply)
	})

	cp.customHandleFunc(helloEvent, func(rw of.ResponseWriter, r *of.Request) {
		//Send back the Hello response

		glog.Infoln("Responded to", of.TypeHello, "from host", r.Addr, ".")

		rw.Write(&of.Header{Type: of.TypeHello}, nil)

		// Features Request
		glog.Infoln("Features Request to ", r.Addr)
		rw.Write(&of.Header{Type: of.TypeFeaturesRequest}, nil)

	})

	cp.customHandleFunc(echoRequestEvent, func(rw of.ResponseWriter, r *of.Request) {
		glog.Infoln("Echo request from", r.Addr, ". Replying.")

		var req ofp.EchoRequest
		req.ReadFrom(r.Body)

		echoReply := of.NewRequest(of.TypeEchoReply, &ofp.EchoReply{
			Data: req.Data,
		})

		rw.Write(&of.Header{Type: of.TypeEchoReply}, echoReply)
	})

	glog.Info("Control plane firing up engines.")

	of.ListenAndServe(fmt.Sprintf(":%d", port), cp.mux)
}

// Stop will kill the control plane listener
func (cp *OpenFlow13ControlPlane) Stop() {

}

// SetupLayer2Switching will cause the controller to instruct devices to behave
// like layer 2 switches.
func (cp *OpenFlow13ControlPlane) SetupLayer2Switching() {
	glog.Infof("Setting up the Layer 2 Switching logic...")

	cp.customHandleFunc(featuresReplyEvent, func(rw of.ResponseWriter, r *of.Request) {
		var featuresReply ofp.SwitchFeatures
		featuresReply.ReadFrom(r.Body)

		// Packets arriving at the ctrlTable will be sent to the controller and then the fwdTable.
		flowModCtrl := ofp.NewFlowMod(ofp.FlowAdd, nil)
		flowModCtrl.Match = ofputil.ExtendedMatch(matchEverything)
		flowModCtrl.Instructions = ofp.Instructions{sendController, gotoForwardingTable}
		flowModCtrl.HardTimeout = 0
		flowModCtrl.Priority = 100
		flowModCtrl.Table = ctrlTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModCtrl)

		// All packets which can't be matched, explicitly flood to all remaining ports.
		flowModCustomMiss := ofp.NewFlowMod(ofp.FlowAdd, nil)
		flowModCustomMiss.Match = ofputil.ExtendedMatch(matchEverything)
		flowModCustomMiss.Instructions = ofp.Instructions{floodPorts}
		flowModCustomMiss.HardTimeout = 0
		flowModCustomMiss.Priority = 100
		flowModCustomMiss.Table = fwdTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModCustomMiss)

		// Example: block all tcp packets via port 80 using a flow mod
		// 	Prerequisites:
		//		OXM_OF_ETH_TYPE in (0x0800, 0x86dd)
		//		OXM_OF_IP_PROTO in (0x06)

		flowModBlockHTTP := ofp.NewFlowMod(ofp.FlowAdd, nil)
		flowModBlockHTTP.Match = ofputil.ExtendedMatch(
			matchEthType0800,
			matchIPProto6,
			matchHTTP,
		)
		flowModBlockHTTP.Instructions = ofp.Instructions{dropPacket}
		flowModBlockHTTP.Priority = 1000
		flowModBlockHTTP.Table = ctrlTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModBlockHTTP)
	})

	cp.customHandleFunc(packetInEvent, func(rw of.ResponseWriter, r *of.Request) {
		var packet ofp.PacketIn
		packet.ReadFrom(r.Body)

		var ingressPort ofp.XMValue

		ingressPort = packet.Match.Field(ofp.XMTypeInPort).Value

		portOutput := &ofp.InstructionApplyActions{
			ofp.Actions{
				&ofp.ActionOutput{ofp.PortNo(ingressPort.UInt32()), 0},
			},
		}

		var packetDecode layers.Ethernet
		packetDecode.DecodeFromBytes(packet.Data, gopacket.NilDecodeFeedback)

		glog.Infof("Learning - Src MAC: %x, Dst MAC: %x", []byte(packetDecode.SrcMAC), []byte(packetDecode.DstMAC))

		matchEthDst := ofp.XM{
			Class: ofp.XMClassOpenflowBasic,
			Type:  ofp.XMTypeEthDst,
			Value: ofp.XMValue(packetDecode.SrcMAC),
		}

		matchEthSrc := ofp.XM{
			Class: ofp.XMClassOpenflowBasic,
			Type:  ofp.XMTypeEthSrc,
			Value: ofp.XMValue(packetDecode.SrcMAC),
		}

		// Add a flow to the fwdTable which matches the packet destination to an output port.
		flowModLearn := ofp.NewFlowMod(ofp.FlowAdd, nil)
		flowModLearn.Match = ofputil.ExtendedMatch(matchEthDst)
		flowModLearn.Instructions = ofp.Instructions{portOutput}
		flowModLearn.HardTimeout = 300
		flowModLearn.Priority = 200
		flowModLearn.Table = fwdTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModLearn)

		// Add a flow to the ctrlTable which matches the packet source in order to avoid
		// sending the packet to the controller if the mapping has already been learned.
		flowModSkipPacketIn := ofp.NewFlowMod(ofp.FlowAdd, nil)
		flowModSkipPacketIn.Match = ofputil.ExtendedMatch(matchEthSrc)
		flowModSkipPacketIn.Instructions = ofp.Instructions{gotoForwardingTable}
		flowModSkipPacketIn.IdleTimeout = 300
		flowModSkipPacketIn.Priority = 200
		flowModSkipPacketIn.Table = ctrlTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModSkipPacketIn)

	})
}
