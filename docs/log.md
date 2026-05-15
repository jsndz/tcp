We can implement TCP in two ways:

1. Upon UDP: on UDP adding reliability and other features of TCP
2. Upon raw IP: you get raw IP data and upon that we are implementing Custom TCP protocol

Let's go with second one.

TCP upon Raw IP.

# 1: Connect to Raw IP 

To get raw IP we need to connect to the Socket which provides raw IP.
So when the IP comes from the network (ipv4), it mentions the protocol number 
what we will do is observe(not steal) the raw IP data.
using syscall pkg

# 2: Sending to socket with protocol

Since we need to test if we can send to protocol create a new process which send to the socket
with same protocol. 
Since there are i have no friend who will take the data give ip addr as 127.0.0.1 which configs to our own system

# 3: Packet structure

Every protocol defines rules.
So each packet is structured in some format both sender and reciever can agree on common format

```go
type Packet struct {
	Version    uint8 // type of version
	SEQ        uint // seq of packet for getting correct order
	ACK        uint // ack bit
	Flags      uint8 // flags are used to combine packet types like sync + ack
	Window     uint16 // window indicates what size of data can the reciever accept
	PayloadLen uint16 
	Checksum   uint16 // checking error
	Payload    []byte
}
```

# 4: Serialization and deserialization of Packet

The packet is sent as byte in the network 
need proper way to serialize and deserialize struct
Using BigEndian for byte order
always give the specific byte size if you are s/d data like uint8,16 
checksum also needs to calculated
used for error detection 
when marshalling keep the checksum as 0 
after marshall and you send the data 
and you unmarshall you can validate the data 
and if no error is there the checksum of the whole packet will be 0
