package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"labrpc"
	"math/rand"
	"sync"
	"time"
	//"fmt"
	"math"
)

// import "bytes"
// import "encoding/gob"

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	// Your data here.
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int
	votedFor    int
	state       int // 0 = follower, 1 = candidate, 2 = leader
	//log
	// Do yo need to do anything with log in a3?
	replyChan         chan *RequestVoteReply
	appendEntriesChan chan *AppendEntriesReply
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here.
	term = rf.currentTerm
	isleader = false
	if rf.state == 2 {
		isleader = true
	}
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here.
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	// Your code here.
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
}

// example RequestVote RPC arguments structure.
type RequestVoteArgs struct {
	// Your data here.
	Term        int // candidate's term
	CandidateId int // candidate requesting vote

	//log?
}

// example RequestVote RPC reply structure.
type RequestVoteReply struct {
	// Your data here.
	Term        int  // currentTerm for candidate to update itself
	VoteGranted bool // true means candidate received vote
	Success bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {

	if rf.state == 0 { // follower
		
		if args.Term < rf.currentTerm {
			reply.Success = false
			reply.Term = rf.currentTerm
			reply.VoteGranted = false
			
		} else if args.Term == rf.currentTerm{
			// follower should not receive from a candidate w equal term
			reply.Success = false 
			reply.VoteGranted = false
		}else {
			//TODO this condition is iffy
			reply.Success = true
			reply.Term = args.Term
			rf.currentTerm = args.Term
			if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
				rf.votedFor = args.CandidateId
				reply.VoteGranted = true
			}
		}
		// necessary so the follower can reset its timer
		rf.replyChan <- reply

	} else if rf.state == 1 { // candidate
		if args.Term < rf.currentTerm {
			reply.Success = false
			reply.Term = rf.currentTerm
			reply.VoteGranted = false
			
		} else if args.Term == rf.currentTerm{
			reply.Success = false 
			reply.VoteGranted = false
		}else {
			reply.Success = true
			reply.Term = args.Term
			rf.votedFor = -1
			if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
				rf.votedFor = args.CandidateId
				reply.VoteGranted = true
			}
			//TODO currently the vote doesn't get sent when demotions happen, should you notify the sender that the vote needs to get resent??
		}

	} else if rf.state == 2 { // leader
		if args.Term < rf.currentTerm {
			reply.Success = false
			reply.Term = rf.currentTerm
			reply.VoteGranted = false
			
		} else if args.Term == rf.currentTerm{
			// leader should not receive from a candidate w equal term
			reply.Success = false 
			reply.VoteGranted = false
		}else {
			//TODO this condition is iffy
			reply.Success = true
			reply.Term = args.Term
			rf.votedFor = -1 //TODO check the resends on this
			if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
				rf.votedFor = args.CandidateId
				reply.VoteGranted = true
			}
		}
		// necessary so leader knows if its term is stale
		rf.replyChan <- reply
	}

	

}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	// does this need to be in a go routine?
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)

	if ok {
		rf.replyChan <- reply
	}

	return ok
}

// student code:
type AppendEntriesArgs struct {
	Term     int // leader's term
	LeaderId int // TODO needed?

}
type AppendEntriesReply struct {
	Term    int // currentTerm for leader to update itself
	Success bool // TODO needed?
	Demote bool
}

// AppendEntries handler
func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {

	if rf.state == 0 { // follower
		if args.Term < rf.currentTerm{
			reply.Success = false
		}else {
			reply.Success = true
			reply.Term = args.Term
			rf.currentTerm = args.Term
		}
		
	} else if rf.state == 1 { // candidate
		if args.Term < rf.currentTerm{
			reply.Success = false
		}else {
			reply.Success = true
			reply.Term = args.Term
		}

	} else if rf.state == 2 { // leader
		if args.Term < rf.currentTerm{
			reply.Success = false
		}else {
			reply.Demote = true
			reply.Term = args.Term
		}
	}
	rf.appendEntriesChan <- reply

}

func (rf *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)

	if ok {
		rf.appendEntriesChan <- reply
	}

	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	return index, term, isLeader
}

// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here.
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.state = 0 // starts off as a follower
	//rf.timer start it
	rf.replyChan = make(chan *RequestVoteReply, 30)           // TODO 30 is arbirtary set this to num peers?
	rf.appendEntriesChan = make(chan *AppendEntriesReply, 30) //TODO 30 is arbitr

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.follower()

	return rf
}

func (rf *Raft) follower() {
	//election timer --random
	rf.state = 0
	rf.votedFor = -1
	random_duration := time.Duration(rand.Intn(300-150) + 150)
	ticker := time.NewTicker(random_duration * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			//follower becomes candidate
			rf.state = 1
			//rf.votedFor = -1
			// rf.currentTerm += 1
			go rf.candidate()
			return
		case appendEntryReply := <-rf.appendEntriesChan:
			if appendEntryReply.Success {
				//resets the timer
				random_duration = time.Duration(rand.Intn(300-150) + 150)
				ticker = time.NewTicker(random_duration * time.Millisecond)
			}
		case RequestVoteReply := <-rf.replyChan:
			if RequestVoteReply.Success{
				// resets the timer
				random_duration = time.Duration(rand.Intn(300-150) + 150)
				ticker = time.NewTicker(random_duration * time.Millisecond)
			}
		}
	}
}

func (rf *Raft) candidate() {
	numVotes := 1 //votes for itself
	numPeers := len(rf.peers)

	//election timer --random
	random_duration := time.Duration(rand.Intn(300-150) + 150)
	ticker := time.NewTicker(random_duration * time.Millisecond)

	//updates value after election timeout and sends RequestVote rpcs to peers
	update_SendRequestVotes := func() {
		// updating values after election timeout
		rf.state = 1
		rf.currentTerm += 1
		rf.votedFor = rf.me
		numVotes = 1
		// sendRequestVote to all peers
		requestVoteArgs := RequestVoteArgs{rf.currentTerm, rf.me}
		reply := &RequestVoteReply{}
		for i := 0; i < numPeers; i++ {
			if i != rf.me {
				go func(i int) {
					ok := rf.sendRequestVote(i, requestVoteArgs, reply)
					// TODO what do you do if the sendRequestVote was not sent?
					if !ok {
						//TODO what needs to happen?
						//TODO Resend rpc???
					}
				}(i)
			}
		}
	}
	update_SendRequestVotes()

	
	for {
		select {
		case <-ticker.C:
			update_SendRequestVotes()
			random_duration = time.Duration(rand.Intn(300-150) + 150)
			ticker = time.NewTicker(random_duration * time.Millisecond)
		case reqVoteReply := <-rf.replyChan:
			// If the request term is less than the current term then it rejects the request
			if reqVoteReply.Term >= rf.currentTerm {
				// if the candidate discovers its term is out of date it becomes a follower
				if reqVoteReply.Term > rf.currentTerm {
					// update term
					rf.currentTerm = reqVoteReply.Term
					// demote to follower
					rf.state = 0
					rf.votedFor = -1
					go rf.follower()
					return
				}
				if reqVoteReply.VoteGranted {
					numVotes += 1
				} //TODO what happens if a vote is not granted?

				//TODO check this condition
				// If candidate has received a majority of the votes make it a leader
				if numVotes >= int(math.Ceil(float64(numPeers)/2.0)) {
					//promote to leader
					rf.state = 2
					go rf.leader()
					return
				}
			}
D
		case appendEntryReply := <-rf.appendEntriesChan:
			if appendEntryReply.Success {
				rf.state = 0
				rf.currentTerm = appendEntryReply.Term
				rf.votedFor = -1
				go rf.follower()
				return
			}

		}

	}
D
}

func (rf *Raft) leader() {
	// remain leader unless you receive an append entries reply
	rf.state = 2
	numPeers := len(rf.peers)

	// TODO heartbeat timer --50 or 100 ms
	ticker := time.NewTicker(50 * time.Millisecond)
	send_heartbeats := func() {
		appendEntriesArgs := AppendEntriesArgs{rf.currentTerm, rf.me}
		reply := &AppendEntriesReply{}
		for i := 0; i < numPeers; i++ {
			if i != rf.me {
				go func(i int) {
					ok := rf.sendAppendEntries(i, appendEntriesArgs, reply)
					if !ok {
						
					}
				}(i)
			}
		}

	}
	send_heartbeats()

	for {
		select {
		case <-ticker.C:
			send_heartbeats()

		case reqVoteReply := <-rf.replyChan:
			if reqVoteReply.Success{
				rf.state = 0
				rf.votedFor = -1
				rf.currentTerm = reqVoteReply.Term
				go rf.follower()
				return
			}
		case appendEntryReply := <-rf.appendEntriesChan:
			if appendEntryReply.Demote {
				rf.state = 0
				rf.votedFor = -1
				rf.currentTerm = appendEntryReply.Term
				go rf.follower()
				return
			}
		}
	}

}
