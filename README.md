# TCP Clone over Raw IP

This project is a custom implementation of a TCP-like protocol built directly on top of Raw IP sockets. It uses protocol number `200` to distinguish its traffic from standard TCP (protocol `6`) or UDP traffic.

## Overview

Implementing a transport protocol over Raw IP allows for full control over packet headers, reliability mechanisms, and flow control. This project demonstrates the core concepts of TCP, including:

- **Raw Socket Integration**: Using `syscall.Socket` with `AF_INET` and `SOCK_RAW` to send and receive custom IP packets.
- **Custom Packet Structure**: A specialized header containing sequence numbers, acknowledgment numbers, flags, and window sizes.
- **Handshake Mechanism**: A SYN/ACK-based connection establishment.
- **Reliability**:
  - **Checksums**: Error detection for packet integrity.
  - **Sequence & Acknowledgment Numbers**: Tracking data delivery and ensuring correct ordering.
  - **Retransmission**: Background loops with exponential backoff for lost packets.
- **Flow Control**: A sliding window implementation to manage "in-flight" data.
- **Multiplexing**: Port-based application differentiation within the custom protocol.

## Project Structure

- `cmd/tcp/`: A sample application demonstrating a client-server chat over the custom protocol.
- `cmd/cli/`: A simple tool for sending raw protocol 200 data.
- `internal/connection/`: Core logic for connection management, handshakes, and retransmission.
- `internal/packet/`: Packet structure, serialization (BigEndian), and validation.
- `internal/util/`: Utility functions, including checksum calculation.

## Packet Structure

Each packet is serialized using BigEndian byte order with the following structure:

| Field | Type | Description |
| :--- | :--- | :--- |
| Version | uint8 | Protocol version |
| Flags | uint8 | Control flags (SYN, ACK, FIN, RST, DATA) |
| SrcPort | uint16 | Source Port |
| DstPort | uint16 | Destination Port |
| SEQ | uint32 | Sequence Number |
| ACK | uint32 | Acknowledgment Number |
| Window | uint16 | Receiver's Window Size |
| PayloadLen | uint16 | Length of the payload |
| Checksum | uint16 | 16-bit checksum for error detection |
| Payload | []byte | The actual data |

## Reliability & Flow Control

### Retransmission
The protocol maintains a `SendBuffer` of unacknowledged packets. A background goroutine periodically checks for timeouts and retransmits packets using exponential backoff (up to 10 retries).

### Out-of-Order Handling
Packets arriving out of order are stored in a `RecvBuffer`. When the missing packet arrives, the protocol "replays" the buffered packets to the application layer in the correct sequence.

### Sliding Window
Transmission is governed by a sliding window. The sender tracks `SendBase` (oldest unacknowledged) and `SendSeq` (next sequence to send). Data is only sent if it falls within the window advertised by the receiver.

## Getting Started

### Prerequisites
- Go 1.x
- **Root/Administrative Privileges**: Raw sockets require elevated permissions to open.

### Running the Demo

1.  **Start the Server**:
    ```bash
    sudo go run cmd/tcp/main.go
    ```
    - Source Port: `8080`
    - Destination Port: `8081`
    - Mode: `server`

2.  **Start the Client**:
    In another terminal:
    ```bash
    sudo go run cmd/tcp/main.go
    ```
    - Source Port: `8081`
    - Destination Port: `8080`
    - Mode: `client`

3.  **Chat**: Type messages in either terminal to see them delivered reliably over protocol 200.

## Implementation Details

- **Protocol Number**: 200 (Custom)
- **Serialization**: `encoding/binary` BigEndian
- **Socket**: `syscall.AF_INET`, `syscall.SOCK_RAW`, `200`
- **Filtering**: Since raw sockets receive all packets for the specified protocol, the implementation includes manual filtering by port and IP address to isolate connections.
