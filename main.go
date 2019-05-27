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
	var cp infrastructure.ControlPlane

<<<<<<< Updated upstream
	helloEvent := of.TypeMatcher(of.TypeHello)
	packetInEvent := of.TypeMatcher(of.TypePacketIn)
	errorEvent := of.TypeMatcher(of.TypeError)

	mux := of.NewServeMux()

	mux.HandleFunc(errorEvent, func(rw of.ResponseWriter, r *of.Request){
		var packet ofp.Error
		packet.ReadFrom(r.Body)

		glog.Errorln("Error:",packet.Error())
	})

	gotoForwardingTable := &ofp.InstructionGotoTable{
		fwdTable,
	}

	mux.HandleFunc(helloEvent, func(rw of.ResponseWriter, r *of.Request){
		//Send back the Hello response

		glog.Infoln("Responded to", of.TypeHello, "from host", r.Addr, ".")

		rw.Write(&of.Header{Type: of.TypeHello}, nil)

		flood := &ofp.InstructionApplyActions{
			ofp.Actions{
				&ofp.ActionOutput{ofp.PortFlood, 0},
			},
		}

		controller := &ofp.InstructionApplyActions{
			ofp.Actions{
				&ofp.ActionOutput{ofp.PortController, ofp.ContentLenMax},
			},
		}

		matchEverything := ofp.XM{
			Class:	ofp.XMClassOpenflowBasic,
			Type:	ofp.XMTypeEthDst,
			Value:	ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Mask:	ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		}

		// Packets arriving at the ctrlTable will be sent to the controller and then the fwdTable.
		FlowModCtrl := ofp.NewFlowMod(ofp.FlowAdd, nil)
		FlowModCtrl.Match = ofputil.ExtendedMatch(matchEverything)
		FlowModCtrl.Instructions = ofp.Instructions{controller, gotoForwardingTable,}
		FlowModCtrl.HardTimeout = 0
		FlowModCtrl.Priority = 100
		FlowModCtrl.Table = ctrlTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, FlowModCtrl)

		// All packets which can't be matched, explicitly flood to all remaining ports.
		FlowModCustomMiss := ofp.NewFlowMod(ofp.FlowAdd, nil)
		FlowModCustomMiss.Match = ofputil.ExtendedMatch(matchEverything)
		FlowModCustomMiss.Instructions = ofp.Instructions{flood}
		FlowModCustomMiss.HardTimeout = 0
		FlowModCustomMiss.Priority = 100
		FlowModCustomMiss.Table = fwdTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, FlowModCustomMiss)

	})

	mux.HandleFunc(packetInEvent, func( rw of.ResponseWriter, r *of.Request){

		glog.Infoln("PacketIn Message from host", r.Addr)

		var packet ofp.PacketIn
		packet.ReadFrom(r.Body)
		
		var ingressPort ofp.XMValue

		ingressPort = packet.Match.Field(ofp.XMTypeInPort).Value

		portOutput := &ofp.InstructionApplyActions{
			ofp.Actions{
				&ofp.ActionOutput{ofp.PortNo(ingressPort.UInt32()), 0},
			},
		}

		//var ethDst []byte
		var packetDecode layers.Ethernet
		packetDecode.DecodeFromBytes(packet.Data, gopacket.NilDecodeFeedback)

		glog.Infof("Src MAC: %x, Dst MAC: %x", []byte(packetDecode.SrcMAC), []byte(packetDecode.DstMAC))

		matchEthDst := ofp.XM{
			Class:	ofp.XMClassOpenflowBasic,
			Type:	ofp.XMTypeEthDst,
			Value:	ofp.XMValue(packetDecode.SrcMAC),
		}

		matchEthSrc := ofp.XM{
			Class:	ofp.XMClassOpenflowBasic,
			Type:	ofp.XMTypeEthSrc,
			Value:	ofp.XMValue(packetDecode.SrcMAC),
		}

		// Add a flow to the fwdTable which matches the packet destination to an output port.
		FlowModLearn := ofp.NewFlowMod(ofp.FlowAdd, nil)
		FlowModLearn.Match = ofputil.ExtendedMatch(matchEthDst)
		FlowModLearn.Instructions = ofp.Instructions{portOutput}
		FlowModLearn.HardTimeout = 30
		FlowModLearn.Priority = 200
		FlowModLearn.Table = fwdTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, FlowModLearn)

		// Add a flow to the ctrlTable which matches the packet source in order to avoid 
		// sending the packet to the controller if the mapping has already been learned.
		FlowModSkipPacketIn := ofp.NewFlowMod(ofp.FlowAdd, nil)
		FlowModSkipPacketIn.Match = ofputil.ExtendedMatch(matchEthSrc)
		FlowModSkipPacketIn.Instructions = ofp.Instructions{gotoForwardingTable}
		FlowModSkipPacketIn.HardTimeout = 30
		FlowModSkipPacketIn.Priority = 200
		FlowModSkipPacketIn.Table = ctrlTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, FlowModSkipPacketIn)

	})

	glog.Info("Starting L2 Switch Controller.")
=======
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
>>>>>>> Stashed changes

	cp.Start(6633)

	glog.Info("Finished.")

	glog.Flush()
}