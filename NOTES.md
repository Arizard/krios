# Krios Notes

## OpenFlow 1.3 Key Domain Concepts

### Introduction

Flows have **instructions**. Packets have **actions**. Flows use **instructions** to modify or execute packet **actions**.

### Flows (Flow Entries)

Each flow entry has the fields:

| Match Fields | Priority | Counters | Instructions | Timeouts | Cookie |
|--------------|----------|----------|--------------|----------|--------|

* **Match Field**s are used to match against packets by ingress port, packet headers (e.g. Ethernet address, IPv4/v6 address + port) and metadata from previous tables.
* **Priority** is used to determine which flow in the table is chosed if there are multiple flow matches in a table.
* **Counters** update when a packet is matched
* **Instructions** to modify the action set
* **Timeouts** define the maximum amount of idle time that a flow persists in a flow table.
* **Cookie** data value chosen by the controller. Could be used by the controller as a form of tagging e.g. to filter flow statistics or to remove particular flows.

A flow's *identity* is composed of it's match fields and priority - each pair of match fields and priority in the table represent a unique flow entry.

The flow entry that wildcards all match fields and has priority 0 is known as the **table-miss flow entry**.

### Instructions

* Instructions (example)
  * Instruction 0 `Meter`
    * `meter-id`
  * Instruction 1 `Apply-Actions`
    * `Increment TTL`
      * `null`
    * ...
    * `Output`
      * Port `FLOOD`
      * MaxLen `0`
  * Instruction n `Goto-Table`
    * `next-table-id`

Instructions are executed when a packet matches a flow entry. These instructions can perform a combination of:
* Changes to the packet
* Changes to the packet action set
* Changes to the pipeline processing (e.g. Send to controller via `PACKET_IN` message)

Common instructions:

* **Meter** directs the packet to the specified meter, where it may be dropped due to metering rules.
* **Apply-Actions** Apply the specific action immediately, without any change to the action set.
* **Clear-Actions** clears all the actions in the action set immediately.
* **Goto-Table** Indicates the next table in the processing pipeline.

There may only be one instruction of each type in the flow entry.

### Action Set

Each packet has an associated action set. Each flow entry can **Clear**, **Write** or **Apply** the action set. Action sets persist between flow tables. **When the instruction set of a flow** has no **Goto-Table** instruction, the packet action set is immediately executed after it exits that flow table.

Action sets contain maximum one action of each type, and are always executed in the following order, regardless of the order of adding the actions.

1. copy TTL inwards
2. pop (apply all pop-tag actions)
3. push-mpls
4. push-pbb
5. push-vlan
6. copy TTL outwards
7. decrement TTL
8. set
9. qos
10. group
11. output

### Flow Table

### Matching

Each packet be compared to each flow table sequentially, until a match is encountered. When encountering a match, counters are updated, action set, packet match/set fields and metadata are updated. From here, the packet may have a **Goto-Table n** instruction, where it will repeat the process until no more **Goto-Table n** instructions are found, then the action set will execute.

If no matches are found, it matches the **table-miss flow entry**. If this does not exist, packet is dropped.

# Appendix

## packet in match

```
Match
    Type: OFPMT_OXM (1)
    Length: 12
    OXM field
        Class: OFPXMC_OPENFLOW_BASIC (0x8000)
        0000 000. = Field: OFPXMT_OFB_IN_PORT (0)
        .... ...0 = Has mask: False
        Length: 4
        Value: 1
    Pad: 00000000
```

## flow mod
```
Match
    Type: OFPMT_OXM (1)
    Length: 12
    OXM field
        Class: OFPXMC_OPENFLOW_BASIC (0x8000)
        0000 000. = Field: OFPXMT_OFB_IN_PORT (0)
        .... ...0 = Has mask: False
        Length: 4
        Value: 1
    Pad: 00000000
```