package main 

type Scheduler interface {
	SelectCandidateNodes()
	Score()
	Pick()
}



