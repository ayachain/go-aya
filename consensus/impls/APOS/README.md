APOS Consensus Mechanism
=========================

   ### Introduce
   &emsp;&emsp;The "APOS" consensus mechanism is realized by optimizing and modifying "Aya" 
   on the basis of the traditional "POS" consensus mechanism. Every opportunity 
   APOS mechanism node will think that it is the correct data. Only when the 
   received data is the same correct time, it will recognize the data and 
   reach consensus. The advantage of this is that if the node itself is wrong, 
   it will be used for other nodes. It becomes a Byzantine node, a helpless 
   and untrustworthy node. Although it will not completely discard all the 
   data of the node, once the data shows signs of bifurcation, the other 
   nodes will immediately terminate execution.
   
   * Step 1.&emsp;WatchDog
   
   * Step 2.&emsp;Signature
   
   * Step 3.&emsp;DataLoader
   
   * Step 4.&emsp;Worker
   
   * Step 5.&emsp;Executor
   
   
   ###Details of each link
   
   ##### 1.WatchDog
   > &emsp;&emsp;1.When data enters, the best first entry, after dealing with the trusted 
   relationship, such as whether the node comes from the node that I often 
   accept data, or the node that I don't trust, and the node that I suspect,
   uses a credit degree to evaluate the trust degree of the node.
   
   > &emsp;&emsp;2.The exception, however, is that if the message comes from a super node, it 
   will be unconditionally trusted, and other nodes will need to be screened to 
   pass the correct message to the next link.
   
   > &emsp;&emsp;3.WachDog also does the initial unpacking of the message, because there are 
   many types of messages, in which we decode the message and continue to deliver 
   it to the object.
   
   
   ##### 2.Signature
   > &emsp;&emsp;When the message arrives, the Hash process is used to verify whether the message 
   is complete or tampered with, and then the message source is authenticated by 
   verifying the signature. If all the messages are passed, the next step will be 
   taken. However, for Block's message, in APOS mechanism, Block's outgoing broadcast
   can only be carried out by super-node, but for Block, only the ID of the sent node 
   needs to be verified, not the next step. Signatures exist, and this process is 
   handled in WatchDog.
   
   > &emsp;&emsp;So if you receive Block's broadcast message in this link, you just need to verify 
   that the content is complete, and then you can go to the next step.
   

   ##### 3.DataLoader
   > &emsp;&emsp;When the above two steps are completed and passed successfully, the relative reliability
    and integrity of the source can be proved. Before the final logic is executed, the data that may need
    to be used should be prepared. Although the whole data block is loaded completely every time in the 
    logic, IPFS DAG Services handles the problem of incremental updating, so if the node is run correctly 
    for a long time, only one time. Incremental data needs to be prepared, and when the data is ready, 
    the next step can be taken.


   #### 4.Worker
   > &emsp;&emsp;The validity of the data has been proved when the message is delivered here. The worker 
   uses a VFS instance and reads and processes the logic of the transaction from the prepared data, but 
   the worker can not write the chain data, but can only read it.
   
   
   #### 5.Executor
   > &emsp;&emsp;The final consistency of transactions responsible for constraining data writing is the 
   last link in the "APOS" mechanism, and the final execution result is the processing result. But the 
   difference is that this link is likely to be single-threaded, because we have not yet thought about 
   how to optimize the process with multi-threading. If we think about it, we will optimize the process.
   
   ### Other
   
   > &emsp;&emsp;In the definition of five stps, the data stored in each link should be a funnel model, 
   because we can think that the first link of each link is simpler than the logic of this link. In order 
   to increase throughput, the number of channel caches is set as follows:
   
    WatchDog    : 256
    Signature   : 64
    DataLoader  : 18
    Worker      : 6
    Executor    : 1