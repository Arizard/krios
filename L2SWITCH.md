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

We need to implement this handshake:

![handshake](http://sdnhub.org/wp-content/uploads/2014/02/OF_Msg_Exchanges.png)

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
