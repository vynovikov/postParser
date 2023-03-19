# DataPiece

DataPiece is the major object in the context of work. It's a pointer to struct containing byte slice and header. PostParser extracts dataPieces from http (or https) request, handles and sends them via gRPC.

## Synchronization and ordering

#### Application

Application uses several methods for dataPiece handling: Presence(), Dec(), Act(), RegisterBuffer(), BufferAdd(). Some of these methods are interacting with store maps. They are executed by workers in concurrent way. Execution needs to be synchronized to avoid data race:



![](forManual/2.png)

sync.RWMutex is used for synchronization

#### Store

DataPieces are reassembled before sending via gRPC. It is important to keep initial order when sending dataPiece group which represent file data chunks. Otherwise file would be corrupted. For that reason store is used:

![](forManual/1.gif)

Store is represented by three maps: Register, Buffer and Counter. 

**Register**

Stores current state of store. If dataPiece's part is matched with Register's, dataPiece is headed to transfer and Register's part is increased by 1, otherwise it stores in Buffer. After successful registration, Buffer elements are trying to register.

**Buffer**

Stores dataPieces. Keeps  being sorted after each addition. Tries to register stored elements after successful registration of new dataPiece

**Counter**

Stores counters for dataPiece groups and flags for marking output as first, last, etc.





## HTTPS
Server listens port 443 for https connections. It uses generated private key and selfsigned certificate. Certificate can be authorized for production purposes.
Key and certificate are in "tls" folder.



## Graceful shutdown

After receiving interrupt signal, application first finishes its current work , then terminates.
![](forManual/3.gif)

​																													\* process durations are shown schematically

#### Action sequence

* HTTP and HTTP listeners are closed immediately.  Application cannot receive new requests from that moment on
* Waiting for receiver goroutines to finish their job, then close chanIn (channel used to deliver new data for application). If there are no jobs, chanIn are closed immediately
* Waiting for application workers to stop, then close chanOut (channel used to deliver data to transmitting module)
* Waiting for transmitter goroutines to stop then close whole app

sync.RWMutex and sync.WaitGroup are used to perform these actions.