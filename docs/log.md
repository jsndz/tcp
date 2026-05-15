We can implement TCP in two ways:

1. Upon UDP: on UDP adding reliability and other features of TCP
2. Upon raw IP: you get raw IP data and upon that we are implementing Custom TCP protocol

Let's go with second one.

TCP upon Raw IP.

Phase 1: Connect to Raw IP 

To get raw IP we need to connect to the Socket which provides raw IP.
So when the IP comes from the network (ipv4), it mentions the protocol number 
what we will do is observe(not steal) the raw IP data.
using syscall pkg