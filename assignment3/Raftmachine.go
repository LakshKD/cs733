package main

import (
	_ "fmt"
	"math"
	"time"
)

/*type error struct {
	err string
}*/

var sm *StateMachine

type Log struct {
	logindex int
	term     int
	entry    []byte
}

type send struct {
	DestID int
	Event  interface{}
}

type LogStore struct {
	index int

	entry []byte
}

type Commit struct {
	index int
	data  []byte
	error string
}

type Alarm struct {
	t int
}

type AppendEv struct {
	data []byte
}
type TimeoutEv struct {
	Time int
}

type AppendEntriesRequestEv struct {
	Term         int
	PrevLogIndex int
	PrevLogTerm  int
	CommitIndex  int
	LeaderId     int
	Entry        []byte
}

type AppendEntriesResponseEv struct {
	Id      int
	Term    int
	Success bool
}

type VoteRequestEv struct {
	Term         int
	LastLogIndex int
	LastLogTerm  int
	CandidateId  int
}

type VoteResponseEv struct {
	Id          int
	Term        int
	VoteGranted bool
}

type StateMachine struct {
	id           int    // server id
	peers        [4]int // other server ids
	term         int
	PrevLogIndex int
	PrevLogTerm  int
	CommitIndex  int
	LastLogIndex int
	LastLogTerm  int
	votedfor     int
	votetracker  [6]int //It tracks from which machine the candidate received the vote
	nextIndex    [6]int
	matchIndex   [6]int
	status       string
	votecounter  int
	Entry        []byte
	LeaderID     int
	count        int
	falsevote    int
	log          [201]Log
}

//AppendEntriesRequest Functions

func (msg *AppendEntriesRequestEv) AppendEntriesRequest_Handler_F() []interface{} {

	var actions = make([]interface{}, 1)

	if sm.term <= msg.Term {

		sm.LeaderID = msg.LeaderId
		sm.term = msg.Term //Update follower's term and set it to leaders term
		sm.votedfor = 0
		if len(msg.Entry) == 0 {

			if sm.CommitIndex < msg.CommitIndex {
				sm.CommitIndex = msg.CommitIndex
				actions = append(actions, Commit{index: msg.CommitIndex, data: sm.log[sm.LastLogIndex].entry})
			}

			actions = append(actions, Alarm{10})
			time.Sleep(2 * time.Second)
			return (actions)

		} else {

			if sm.LastLogIndex == msg.PrevLogIndex && sm.LastLogTerm == msg.PrevLogTerm {
				index := sm.LastLogIndex

				if len(sm.log[index+1].entry) != 0 && sm.log[index+1].term != 0 {

					for i := index + 1; len(sm.log[i].entry) != 0 && sm.log[i].term != 0 && i <= 200; i++ {
						sm.log[i].logindex = 0
						sm.log[i].term = 0

					}

				}

				sm.LastLogIndex++

				if msg.CommitIndex > sm.CommitIndex {

					sm.CommitIndex = int(math.Min(float64(msg.CommitIndex), float64(sm.LastLogIndex)))

				}
				actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, true}})
				actions = append(actions, LogStore{sm.LastLogIndex, msg.Entry})

				if msg.CommitIndex > sm.CommitIndex {

					sm.CommitIndex = int(math.Min(float64(msg.CommitIndex), float64(sm.LastLogIndex)))

				}
				time.Sleep(2 * time.Second)
				return (actions)

			} else if msg.PrevLogIndex != -1 && len(sm.log[msg.PrevLogIndex].entry) == 0 { //no entry at PrevlogIndex(sent by the leader)

				actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, false}})
				time.Sleep(2 * time.Second)
				return (actions)

			} else if msg.PrevLogIndex == -1 {

				if len(sm.log[0].entry) != 0 && sm.log[0].term != 0 {

					for i := 0; len(sm.log[i].entry) != 0 && sm.log[i].term != 0 && i <= 200; i++ {
						sm.log[i].logindex = 0
						sm.log[i].term = 0
						//log[i].entry=nil

					}
					sm.LastLogIndex = 0
					sm.LastLogTerm = msg.Term
				}

				actions = append(actions, LogStore{0, msg.Entry})
				actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, true}})
				time.Sleep(2 * time.Second)
				return (actions)
			} else if sm.LastLogIndex == msg.PrevLogIndex && sm.LastLogTerm != sm.LastLogTerm {

				actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, false}})
				time.Sleep(2 * time.Second)
				return (actions)
			}

		}

	} else {

		actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, false}})
		time.Sleep(2 * time.Second)
		return (actions)
	}
	time.Sleep(2 * time.Second)

	return (make([]interface{}, 1))

}

func (msg *AppendEntriesRequestEv) AppendEntriesRequest_Handler_C() []interface{} {

	var actions = make([]interface{}, 1)

	if sm.term <= msg.Term {
		sm.term = msg.Term
		sm.votedfor = 0
		sm.status = "Follower"
		if len(msg.Entry) == 0 {

			actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, true}})
			return (actions)
		} else {

			actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, false}})
			return (actions)

		}
	} else {

		actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, false}})
		return (actions)

	}
}

func (msg *AppendEntriesRequestEv) AppendEntriesRequest_Handler_L() []interface{} {

	var actions = make([]interface{}, 1)

	if sm.term >= msg.Term {

		actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, false}})

	} else {
		actions = append(actions, send{msg.LeaderId, AppendEntriesResponseEv{sm.id, sm.term, false}})
		sm.term = msg.Term
		sm.votedfor = 0
		sm.status = "Follower"
	}
	return (actions)
}

//AppendEntriesResponse Functions

func (msg *AppendEntriesResponseEv) AppendEntriesReponse_Handler_F() []interface{} {

	var actions = make([]interface{}, 1)
	if sm.term < msg.Term {
		sm.term = msg.Term
		sm.votedfor = 0
	}

	//actions = append(actions, error{"Error"})

	return (actions)

}

func (msg *AppendEntriesResponseEv) AppendEntriesReponse_Handler_C() []interface{} {

	var actions = make([]interface{}, 1)
	if sm.term < msg.Term {
		sm.term = msg.Term
		sm.votedfor = 0
		sm.status = "Follower"
	}

	//actions = append(actions, error{"Error"})

	return (actions)

}

func (msg *AppendEntriesResponseEv) AppendEntriesReponse_Handler_L() []interface{} {

	var actions = make([]interface{}, 1)
	if msg.Success == false {
		//fmt.Println("Hello its false")
		if sm.term < msg.Term {
			sm.term = msg.Term
			sm.votedfor = 0
			sm.status = "Follower"
		} else {

			if sm.nextIndex[msg.Id] > 0 {
				sm.nextIndex[msg.Id]--
			}
			j := sm.nextIndex[msg.Id]
			if j > 0 {
				j = j - 1
			}
			//fmt.Println("Its j",j)
			actions = append(actions, send{msg.Id, AppendEntriesRequestEv{sm.term, j, sm.log[j].term, sm.CommitIndex, sm.id, sm.log[j].entry}})
			//time.Sleep(2 * time.Second)
			return (actions)
		}

	} else if msg.Success == true {
		//fmt.Println("Hello its true")
		sm.matchIndex[msg.Id]++
		j := sm.nextIndex[msg.Id]
		sm.nextIndex[msg.Id]++
		for i := 1; i <= 5; i++ {
			if sm.matchIndex[i] >= j {
				sm.count++
			}
		}
		if sm.count >= 3 {

			//fmt.Println("Hello its trueeeeeeeeeeeeee")
			sm.count = 0
			sm.CommitIndex = sm.LastLogIndex
			actions = append(actions, Commit{index: j, data: sm.log[sm.LastLogIndex].entry})
			for i := 0; i <= 3; i++ {
				actions = append(actions, send{sm.peers[i], AppendEntriesRequestEv{sm.term, sm.PrevLogIndex, sm.PrevLogTerm, sm.CommitIndex, sm.id, sm.Entry}})

			}
			return (actions)
		}
		time.Sleep(2 * time.Second)
	}
	time.Sleep(2 * time.Second)

	return (actions)

}

//VoteRequest Functions

func (msg *VoteRequestEv) VoteRequest_Handler_F() []interface{} {

	var actions = make([]interface{}, 1)
	if msg.Term < sm.term {

		actions = append(actions, send{msg.CandidateId, VoteResponseEv{Id: sm.id, Term: sm.term, VoteGranted: false}})

		return (actions)

	} else if (sm.votedfor == 0 || sm.votedfor == msg.CandidateId) && msg.LastLogTerm >= sm.LastLogTerm && msg.LastLogIndex >= sm.LastLogIndex {

		sm.votedfor = msg.CandidateId
		//fmt.Println(sm.id,"Votedfor",sm.votedfor)
		sm.term = msg.Term
		actions = append(actions, send{msg.CandidateId, VoteResponseEv{Id: sm.id, Term: sm.term, VoteGranted: true}})
		return (actions)
	} else {
		actions = append(actions, send{msg.CandidateId, VoteResponseEv{Id: sm.id, Term: sm.term, VoteGranted: false}})
		return (actions)
	}

}

func (msg *VoteRequestEv) VoteRequest_Handler_C() []interface{} {

	var actions = make([]interface{}, 1)
	if msg.Term < sm.term {
		actions = append(actions, send{msg.CandidateId, VoteResponseEv{Id: sm.id, Term: sm.term, VoteGranted: false}})
		return (actions)
	} else if (sm.votedfor == 0 || sm.votedfor == msg.CandidateId) && msg.LastLogIndex >= sm.LastLogIndex && msg.LastLogTerm >= sm.LastLogTerm {
		if sm.votedfor == 0 {
			sm.term = msg.Term
			sm.votedfor = msg.CandidateId

		}
		actions = append(actions, send{msg.CandidateId, VoteResponseEv{Id: sm.id, Term: sm.term, VoteGranted: true}})
		return (actions)

	} else {

		actions = append(actions, send{msg.CandidateId, VoteResponseEv{Id: sm.id, Term: sm.term, VoteGranted: false}})
		return (actions)
	}

}
func (msg *VoteRequestEv) VoteRequest_Handler_L() []interface{} {

	var actions = make([]interface{}, 1)
	if sm.term < msg.Term {
		sm.term = msg.Term
		sm.votedfor = 0
		sm.status = "Follower"
	}
	//actions = append(actions, error{"Error"})

	return (actions)
}

//VoteResponse functions

func (msg *VoteResponseEv) VoteResponse_Handler_F() []interface{} {

	var actions = make([]interface{}, 1)
	if sm.term < msg.Term {
		sm.term = msg.Term
		sm.votedfor = 0
	}
	//actions = append(actions, error{"Error"})

	return (actions)

}

func (msg *VoteResponseEv) VoteResponse_Handler_C() []interface{} {
	//fmt.Println("INSIDE VOTEE RESP CANDIDATE")
	var actions = make([]interface{}, 1)

	if msg.VoteGranted == true {

		sm.votecounter++
		sm.votetracker[msg.Id] = 1

		if sm.votecounter >= 3 {
			//fmt.Println("Hello am the leader",sm.id)
			sm.status = "Leader"
			sm.LeaderID = sm.id
			for i := 0; i <= 3; i++ {
				actions = append(actions, send{sm.peers[i], AppendEntriesRequestEv{sm.term, sm.PrevLogIndex, sm.PrevLogTerm, sm.CommitIndex, sm.id, sm.Entry}})

			}
			return (actions)
		}
	} else if msg.VoteGranted == false {

		if sm.term < msg.Term {
			sm.term = msg.Term
			sm.votedfor = 0
			sm.status = "Follower"
		}
		sm.falsevote++
		if sm.falsevote >= 3 {
			sm.votedfor = 0
			sm.status = "Follower"
		}

	}
	return (make([]interface{}, 1))

}

func (msg *VoteResponseEv) VoteResponse_Handler_L() []interface{} {

	var actions = make([]interface{}, 1)
	if sm.term < msg.Term {
		sm.term = msg.Term
		sm.votedfor = 0
		sm.status = "Follower"
	}
	//actions = append(actions, error{"Error"})
	return (actions)

}

//Append functions

func (msg *AppendEv) Append_Handler_F() []interface{} {

	var actions = make([]interface{}, 1)

	actions = append(actions, Commit{data: msg.data, error: "NOT a Leader"})

	return (actions)

}

func (msg *AppendEv) Append_Handler_C() []interface{} {

	var actions = make([]interface{}, 1)

	actions = append(actions, Commit{data: msg.data, error: "NOT a Leader"})

	return (actions)

}

func (msg *AppendEv) Append_Handler_L() []interface{} {

	var actions = make([]interface{}, 1)
	//fmt.Println(sm)
	//if(sm.LastLogIndex!=0){
	//if(sm.PrevLogIndex==-1 && sm.LastLogIndex==-1){
	sm.PrevLogIndex++
	sm.LastLogIndex++
	//}else {
	//	sm.LastLogIndex++
	//	sm.PrevLogIndex=sm.LastLogIndex-1
	//}
	sm.matchIndex[sm.id]++
	//fmt.Println(sm.peers[0:])
	for i := 0; i <= 3; i++ {
		sm.nextIndex[sm.peers[i]] = sm.LastLogIndex
	}
	actions = append(actions, LogStore{sm.LastLogIndex, msg.data})
	for i := 0; i <= 3; i++ {
		actions = append(actions, send{sm.peers[i], AppendEntriesRequestEv{sm.term, sm.PrevLogIndex, sm.log[sm.PrevLogIndex].term, sm.CommitIndex, sm.id, msg.data}})

	}
	return (actions)
}

//Timeout funtcions

func (msg *TimeoutEv) Timeout_Handler_F() []interface{} {

	var actions = make([]interface{}, 1)
	s := [6]int{0, 0, 0, 0, 0, 0}
	sm.votetracker = s
	sm.status = "Candidate"
	sm.votedfor = 0
	sm.term++
	sm.votetracker[sm.id] = 1
	sm.votecounter++
	for i := 0; i <= 3; i++ {
		actions = append(actions, send{sm.peers[i], VoteRequestEv{sm.term, sm.LastLogIndex, sm.LastLogTerm, sm.id}})

	}

	return (actions)
}

func (msg *TimeoutEv) Timeout_Handler_C() []interface{} {

	var actions = make([]interface{}, 1)
	s := [6]int{0, 0, 0, 0, 0, 0}
	sm.votetracker = s
	sm.votecounter = 0
	sm.status = "Candidate"
	sm.votedfor = 0
	sm.term++
	sm.votetracker[sm.id] = 1
	sm.votecounter++
	for i := 0; i <= 3; i++ {
		actions = append(actions, send{sm.peers[i], VoteRequestEv{sm.term, sm.LastLogIndex, sm.LastLogTerm, sm.id}})

	}
	return (actions)

}

func (msg *TimeoutEv) Timeout_Handler_L() []interface{} {

	var actions = make([]interface{}, 1)
	//Send HeartBeat message to all the followers
	Entry := []byte{}
	for i := 0; i <= 3; i++ {
		actions = append(actions, send{sm.peers[i], AppendEntriesRequestEv{sm.term, sm.PrevLogIndex, sm.LastLogIndex, sm.CommitIndex, sm.id, Entry}})
	}

	return (actions)

}

func (SM *StateMachine) ProcessEvent(ev interface{}) []interface{} {
	sm = SM

	switch ev.(type) {

	case AppendEntriesRequestEv:
		AEReqEv := ev.(AppendEntriesRequestEv)
		//fmt.Println(SM.id)
		//fmt.Println(SM.status)
		//fmt.Println(AEReqEv)
		if SM.status == "Follower" {
			//fmt.Println("Hello inside Append")
			//fmt.Println("Hey",SM.id,AEReqEv.AppendEntriesRequest_Handler_F())
			return (AEReqEv.AppendEntriesRequest_Handler_F())
		} else if SM.status == "Candidate" {
			return (AEReqEv.AppendEntriesRequest_Handler_C())
		} else if SM.status == "Leader" {
			return (AEReqEv.AppendEntriesRequest_Handler_L())
		}

	case AppendEntriesResponseEv:
		AEResEv := ev.(AppendEntriesResponseEv)
		if SM.status == "Follower" {
			return (AEResEv.AppendEntriesReponse_Handler_F())
		} else if SM.status == "Candidate" {
			return (AEResEv.AppendEntriesReponse_Handler_C())
		} else if SM.status == "Leader" {
			return (AEResEv.AppendEntriesReponse_Handler_L())
		}
	case VoteRequestEv:
		VReqEv := ev.(VoteRequestEv)
		if SM.status == "Follower" {
			//fmt.Println("Inside")
			//fmt.Println(SM.id)
			//fmt.Println(VReqEv.VoteRequest_Handler_F())
			return (VReqEv.VoteRequest_Handler_F())
		} else if SM.status == "Candidate" {
			return (VReqEv.VoteRequest_Handler_C())
		} else if SM.status == "Leader" {
			return (VReqEv.VoteRequest_Handler_L())
		}

	case VoteResponseEv:
		VResEv := ev.(VoteResponseEv)
		if SM.status == "Follower" {
			return (VResEv.VoteResponse_Handler_F())
		} else if SM.status == "Candidate" {
			return (VResEv.VoteResponse_Handler_C())
		} else if SM.status == "Leader" {
			return (VResEv.VoteResponse_Handler_L())
		}

	case AppendEv:
		ApndEv := ev.(AppendEv)
		if SM.status == "Follower" {
			//fmt.Println(ApndEv)
			return (ApndEv.Append_Handler_F())
		} else if SM.status == "Candidate" {
			return (ApndEv.Append_Handler_C())
		} else if SM.status == "Leader" {
			//fmt.Println(sm)
			return (ApndEv.Append_Handler_L())
		}

	case TimeoutEv:
		TimeEv := ev.(TimeoutEv)
		if SM.status == "Follower" {
			//fmt.Println("INSIDE TIMEOUT")

			return (TimeEv.Timeout_Handler_F())
		} else if SM.status == "Candidate" {
			return (TimeEv.Timeout_Handler_C())
		} else if SM.status == "Leader" {
			return (TimeEv.Timeout_Handler_L())
		}

		//default: return (make([]interface{}, 1))
	}

	return (make([]interface{}, 1))

}