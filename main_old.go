package main

import (
	//bytes"
	"flag"
	"github.com/golang/glog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	of "github.com/netrack/openflow"
	"github.com/netrack/openflow/ofp"
	"github.com/netrack/openflow/ofputil"
	// "encoding/binary"
)

const (
	// Packets will enter ctrlTable first
	ctrlTable ofp.Table = ofp.Table(0)
	fwdTable  ofp.Table = ofp.Table(1)
)

func main() {
	flag.Parse()

	helloEvent := of.TypeMatcher(of.TypeHello)
	packetInEvent := of.TypeMatcher(of.TypePacketIn)
	errorEvent := of.TypeMatcher(of.TypeError)
	echoRequestEvent := of.TypeMatcher(of.TypeEchoRequest)
	featuresReplyEvent := of.TypeMatcher(of.TypeFeaturesReply)

	mux := of.NewServeMux()

	mux.HandleFunc(errorEvent, func(rw of.ResponseWriter, r *of.Request) {
		var packet ofp.Error
		packet.ReadFrom(r.Body)

		glog.Errorln("Error:", packet.Error())
	})

	gotoForwardingTable := &ofp.InstructionGotoTable{
		fwdTable,
	}

	controller := &ofp.InstructionApplyActions{
		ofp.Actions{
			&ofp.ActionOutput{ofp.PortController, ofp.ContentLenMax},
		},
	}

	flood := &ofp.InstructionApplyActions{
		ofp.Actions{
			&ofp.ActionOutput{ofp.PortFlood, 0},
		},
	}

	dropPacket := &ofp.InstructionApplyActions{
		ofp.Actions{
			//Empty actions means we drop the packet.
		},
	}

	mux.HandleFunc(featuresReplyEvent, func(rw of.ResponseWriter, r *of.Request) {
		var featuresReply ofp.SwitchFeatures
		featuresReply.ReadFrom(r.Body)

		glog.Infof("Features Reply from %s: DatapathID %x, %v\n", r.Addr, featuresReply.DatapathID, featuresReply)

		matchEverything := ofp.XM{
			Class: ofp.XMClassOpenflowBasic,
			Type:  ofp.XMTypeEthDst,
			Value: ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Mask:  ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		}

		// Packets arriving at the ctrlTable will be sent to the controller and then the fwdTable.
		flowModCtrl := ofp.NewFlowMod(ofp.FlowAdd, nil)
		flowModCtrl.Match = ofputil.ExtendedMatch(matchEverything)
		flowModCtrl.Instructions = ofp.Instructions{controller, gotoForwardingTable}
		flowModCtrl.HardTimeout = 0
		flowModCtrl.Priority = 100
		flowModCtrl.Table = ctrlTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModCtrl)

		// All packets which can't be matched, explicitly flood to all remaining ports.
		flowModCustomMiss := ofp.NewFlowMod(ofp.FlowAdd, nil)
		flowModCustomMiss.Match = ofputil.ExtendedMatch(matchEverything)
		flowModCustomMiss.Instructions = ofp.Instructions{flood}
		flowModCustomMiss.HardTimeout = 0
		flowModCustomMiss.Priority = 100
		flowModCustomMiss.Table = fwdTable

		rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModCustomMiss)

		// Example: block all tcp packets via port 80 using a flow mod
		// 	Prerequisites:
		//		OXM_OF_ETH_TYPE in (0x0800, 0x86dd)
		//		OXM_OF_IP_PROTO in (0x06)

		matchEthType0800 := ofp.XM{
			Class: ofp.XMClassOpenflowBasic,
			Type:  ofp.XMTypeEthType,
			Value: ofp.XMValue{0x08, 0x00},
		}

		matchIPProto6 := ofp.XM{
			Class: ofp.XMClassOpenflowBasic,
			Type:  ofp.XMTypeIPProto,
			Value: ofp.XMValue{0x06},
		}
		matchHTTP := ofp.XM{
			Class: ofp.XMClassOpenflowBasic,
			Type:  ofp.XMTypeTCPDst,
			Value: ofp.XMValue{0x00, 0x50}, // 80 in base 16
		}

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

	mux.HandleFunc(helloEvent, func(rw of.ResponseWriter, r *of.Request) {
		//Send back the Hello response

		glog.Infoln("Responded to", of.TypeHello, "from host", r.Addr, ".")

		rw.Write(&of.Header{Type: of.TypeHello}, nil)

		// Features Request
		glog.Infoln("Features Request to ", r.Addr)
		rw.Write(&of.Header{Type: of.TypeFeaturesRequest}, nil)

	})

	mux.HandleFunc(echoRequestEvent, func(rw of.ResponseWriter, r *of.Request) {
		glog.Infoln("Echo request from", r.Addr, ". Replying.")

		var req ofp.EchoRequest
		req.ReadFrom(r.Body)

		echoReply := of.NewRequest(of.TypeEchoReply, &ofp.EchoReply{
			Data: req.Data,
		})

		rw.Write(&of.Header{Type: of.TypeEchoReply}, echoReply)
	})

	mux.HandleFunc(packetInEvent, func(rw of.ResponseWriter, r *of.Request) {

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

	glog.Info("Starting L2 Switch Controller.")

	of.ListenAndServe(":6633", mux)

	glog.Info("Started!")

	glog.Flush()
}
