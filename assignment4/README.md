Name:- Lakshya Kumar
Roll Number:- 153050051

Problem Definition

In this assignment we have to combine a simple Network Filesystem Server with the RaftNode so that the clients will be able to send the commands like read, write, compare & swap(cas), delete and all the changes because of these commands will get replicated to all the servers using Raft consensus Algorithm.



Code Description:- 

Only the Improved version of RaftNode is present and the rest of the code that will link the filesystem with the RaftNode is under Development.

In assignment 3 I have not implemented the heartbeat timeout and Election Timeout properly so now in this code heartbeat and election timeout functionality is fixed and now the Leader elections are successfully going on and one raft node will become the Leader and after that one append is also performed on the leader.

Note:- In the RaftNode code also there is some problem with the Multiple Append request. When trying to do multiple appends the system is showing some random behaviour that is what I am trying to fix.

How To run the code :- 

just run command in Terminal
go test 


 



