# Writing a Layer 2 Switch using OpenFlow 1.3

## Requirements

* Mininet VM
* Go and Go packages
  * `google/gopacket`
  * `netrack/openflow`
  * `golang/glog`

## Assumptions

* Mininet VM installed
* Go installed
* Go packages installed

## Algorithm Overview

### Initialisation (Hello)

When a switch connects to a controller, the switch immediately sends an OpenFlow Hello message. The controller will response with a Hello Response. This is used to negotiate OpenFlow version support.

After the Hello Response is sent, our controller sends two flows:

1. A flow entry on table 0, which matches all packets, with priority 100, with these instructions:
    1. Apply these actions:
       * Forward to the controller
    1. Go to table 1
2. A flow entry on table 1, which matches all packets, with priority 100, with these instructions:
    1. Apply these actions:
       * Forward on all switch ports except the incoming port

Given this information only, the switch will first send the packet to the controller, then flood the packet on all ports.

### PacketIn

This works for now, but it is slow and introduces network congestion - all nodes connected to the switch will exist in the same collision domain.

To improve our switch, and make it 'learn' the location of each node, we must implement the following:

Every time a packet is sent to the controller, we would like to map the source address of the packet to the ingress switch port of the packet.

To achieve this, we listen for the PacketIn event, and then add a new flow on table 1 which matches the destination address of a packet to the source address of the packet we just recieved, and forwards the packet to the switch port we just received the packet from.

| Packet Headers |     | Flow Entry |
| -------------- | --- | ---------- |
| SrcMAC         | →   | DstMAC     |
| InPort         | →   | OutPort    |

Every time a packet is sent from switch to controller, it is encapsulated by the OpenFlow PacketIn message. Using the information from the PacketIn message, we can devise a procedure:

1. Listen for the PacketIn message.
2. Receive PacketIn.
   1. Copy the source MAC address (`ethSrc`)
   2. Copy the ingress port (`inPort`)
3. Create a new entry on table 1, which matches packets where destination MAC address is `ethSrc`, with priority **200**, with these instructions
   1. Apply these actions:
        * Forward on `inPort`

With this configuration, whenever a packet arrives on the switch, the packet is sent to the controller and the node location is 'learned' by the switch.

This is still not enough - it's poor design to send EVERY packet to the switch, especially since we will often already know the location of a node.

Let's add one more flow entry, on the same PacketIn event.

4. Create a new entry on table 0, which matches packets where **source MAC address** is `ethSrc`, with priority **200**, with these instructions:
   1. Go to table 1

Notice the **priority**? Matches are sorted by descending priority. If multiple matches are possible, the switch chooses the one with highest priority. This ensures that we will skip the flow entry that forwards packets to the switch, if the node it comes from has already been learned.

## Go Implementation

Note that this is an example, and won't compile on its own.

### Initialisation (Hello)

```go
helloEvent := of.TypeMatcher(of.TypeHello)

mux := of.NewServeMux()

gotoForwardingTable := &ofp.InstructionGotoTable{
fwdTable,
}
```

```go
mux.HandleFunc(helloEvent, func(rw of.ResponseWriter, r *of.Request){

    //Send back the Hello response
    glog.Infoln("Responded to", of.TypeHello, "from host", r.Addr, ".")

    rw.Write(&of.Header{Type: of.TypeHello}, nil)

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

    matchEverything := ofp.XM{
        Class:  ofp.XMClassOpenflowBasic,
        Type:   ofp.XMTypeEthDst,
        Value:  ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
        // A mask of 6 0x00 bytes tells the switch that we don't actually care about any
        // bits in the ethernet address - this will always match.
        Mask:   ofp.XMValue{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
    }

    // Packets arriving at the ctrlTable will be sent to the controller and then
    // the fwdTable.
    flowModCtrl := ofp.NewFlowMod(ofp.FlowAdd, nil)
    flowModCtrl.Match = ofputil.ExtendedMatch(matchEverything)
    flowModCtrl.Instructions = ofp.Instructions{controller, gotoForwardingTable,}
    flowModCtrl.Priority = 100
    flowModCtrl.Table = ctrlTable

    rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModCtrl)

    // All packets which can't be matched, explicitly flood to all remaining ports.
    flowModCustomMiss := ofp.NewFlowMod(ofp.FlowAdd, nil)
    flowModCustomMiss.Match = ofputil.ExtendedMatch(matchEverything)
    flowModCustomMiss.Instructions = ofp.Instructions{flood}
    flowModCustomMiss.Priority = 100
    flowModCustomMiss.Table = fwdTable

    rw.Write(&of.Header{Type: of.TypeFlowMod}, flowModCustomMiss)

})
```

### PacketIn

```go
packetInEvent := of.TypeMatcher(of.TypePacketIn)
```

```go
mux.HandleFunc(packetInEvent, func( rw of.ResponseWriter, r *of.Request){

    glog.Infoln("PacketIn Message from host", r.Addr)

    var packet ofp.PacketIn
    packet.ReadFrom(r.Body)

    var ingressPort ofp.XMValue

    ingressPort = packet.Match.Field(ofp.XMTypeInPort).Value

    // Instruction to immediately output the packet on the ingress port
    portOutput := &ofp.InstructionApplyActions{
        ofp.Actions{
            &ofp.ActionOutput{ofp.PortNo(ingressPort.UInt32()), 0},
        },
    }

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
    flowModLearn.IdleTimeout = 300
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
```