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

Assume a Switch with a 4 ethernet ports (switch ports). The switch initial state has a single flow table (Table 0) and a single flow entry (table-miss).

A packet arrives from **host 1** on **port 1**, destined for **host 2**. It is unknown which port **host 2** is connected on, or even if **host 2** exists.

**table-miss** is the first matching flow found in Table 0. The packet is sent to the controller via **PacketIn** message. 

The controller creates a **FlowModification** message with the following information:

| Packet Field  | Value                             |
|---            |---                                |
| Table         | 0                                 |
| Command       | ADD                               |
| IdleTimeout   | 0                                 |
| HardTimeout   | 300                               |
| Priority      | 1000                              |
| Buffer        | (Buffer uint32 from PacketIn)     |
| Flags         | (bit flags - not elaborated here) |
| Match         | A list of fields to match         |
| Instructions  | A list of instructions            |

Let's expand the **Match** fields:

| Match Field   | Value                             |
|---            |---                                |
| EthAddrDst    | <EthAddr host 1>                  |

Let's expand the **Instructions** fields:

| Instruction   | Value                                         |
|---            |---                                            |
| Apply-Action  | ActionSet[ Output{ Port: <Switch Port 1>, MaxLen: 0 } ] |

The above **FlowModification** tells the switch to create a new flow entry in Table 0. The flow entry will match against any packet arriving on any switch port, where the Ethernet address is identical to the address of **host 1**.

Once a match is encountered, the switch processed each **Instruction** in sequence. **Apply-Action** is an instruction which directs the switch to immediately apply the action set it provides. In this case, the switch will immediately forward the packet on switch port 1.

Note the **HardTimeout** field - the flow entry will expire and be removed after 300 seconds (5 minutes).

Note the **Buffer** field - this is 32 bits which identify the packet buffered on the switch. This tells the switch which packet this FlowModification is related to.

However, this is not enough - we still need to forward the packet onwards. Since we don't know where **host 2** is, we must forward on all ports except the ingress port.

**PacketOut** needs the following information:

| Packet Field      | Value                         |
|---                |---                            |
| Buffer            | (Buffer from PacketIn)        |
| InPort            | (Ingress Port from PacketIn)  |
| Actions           | ActionSet[ Output{ Port: FLOOD, MaxLen: 0 }] |

This results in the packet from PacketIn (using Buffer to identify the packet) being forwarded on all ports (a 'flood').

The expected behaviour is that the destination node will send a response, and 