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


# 5: SEQ and ACK

tracking seq is simple as adding length of the data to seq
holding the sendSeq and RecvSeq 
helps to know what data has been transmitted
seq is len of data recv or sent
ack are sent for every packet
if there is data to be sent by application layer
then combine that and send ack with data 
that is reciever response
DATA consumes sequence space by payload length
SYN consumes 1 sequence number
FIN consumes 1 sequence number


# 6: Retranmission Logic

So retransmission logic is always running in the background
so if there are any packet without ack in the conn's sendBuffer
it retries here we are doing exponential backoff
if the packet receives ack then it is removed
for example if the packet 1005 recieves ack
then packet that have less seq than it will all be ack 

like you recv a seq lets say 1000
then you if you have any seq in the sendbuffer below 1000 those have been ack 
the ack will be like like 1005
so all the seq below 1005 => seq+ len(sent.payload)
and you can delete them now


# 7: Out of order packet handling

if the packets are out of order just store them in the recv buffer
now when you are recv the packet 
if you next seq is in the recv buffer then use it and then handle it
here handling means putting in chan

Out of order packet arrives store in the recv buffer
when the correct packet arrives check for buffer if that packet exist in the buffer process it 
and repeat till the seq becomes recv eseq


# 8:  Add duplicate detection

send ack and return
because the ack prev sent might have been lost


# 9: Sliding window 

The current setup is more of a stop and wait.
You have to wait around for the data ack to continue.
Next step would be adding a layer on top where 
you have multiple packets to send.

Imagine it like this you are continously sending multiple data and you dont wait around the sent data is in the network
but there is a limit to what can be in the air (called "in flight")that is the window size
as soon as the window size gets full it stops 

and after getting ack then you can send more

so the window slides
```go
type Connection struct {
	mu sync.Mutex

	SocketFD int

	SendSeq    uint32
	SendBase   uint32 // oldest unacknowledged SEQ
	SendWindow uint32 // max number of bytes that can be sent without ACK
	RecvSeq    uint32
	RecvWindow uint32 // max number of bytes that can be received without ACK -> your window size

	PeerAddr *syscall.SockaddrInet4

	SendBuffer map[uint32]*SendPacket
	RecvBuffer map[uint32]*packet.Packet

	DeliveryChan chan *packet.Packet
}
```
The  sliding-window transmission policy is usually one layer above the raw Send() primitive.
