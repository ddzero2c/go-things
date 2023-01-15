package main

import (
	"log"
)

type state int

const (
	A state = iota
	B
)

type stateActor interface {
	A()
	B()
}

type stateA struct {
	sm *stateMachine
}

func (a *stateA) A() {
	a.sm.logger.Printf("state A -> B")
	a.sm.state = B
}

func (a *stateA) B() {
	a.sm.logger.Printf("state A -> do nothing")
}

type stateB struct {
	sm *stateMachine
}

func (b *stateB) A() {
	b.sm.logger.Printf("state B -> do nothing")
}

func (b *stateB) B() {
	b.sm.logger.Printf("state B -> A")
	b.sm.state = A
}

type stateMachine struct {
	state   state
	actors  []stateActor
	actionc chan func()
	quitc   chan chan struct{}
	logger  *log.Logger
}

func newStateMachine() *stateMachine {
	sm := &stateMachine{
		actionc: make(chan func()),
		quitc:   make(chan chan struct{}),
	}
	sm.actors = []stateActor{
		&stateA{sm: sm},
		&stateB{sm: sm},
	}
	sm.logger = log.New(log.Writer(), "state machine: ", log.Flags())
	go sm.loop()
	return sm
}

func (sm *stateMachine) loop() {
	sm.logger.Printf("state machine started")
	defer func() { sm.logger.Printf("state machine stopped") }()
	for {
		select {
		case fn := <-sm.actionc:
			fn()
		case q := <-sm.quitc:
			close(q)
			return
		}
	}
}

func (sm *stateMachine) close() {
	sm.logger.Printf("trigger close")
	q := make(chan struct{})
	sm.quitc <- q
	<-q
	sm.logger.Printf("closed")
}

func (sm *stateMachine) A() {
	sm.actionc <- func() { sm.actors[sm.state].A() }
}

func (sm *stateMachine) B() {
	sm.actionc <- func() { sm.actors[sm.state].B() }
}

func main() {
	sm := newStateMachine()
	sm.A()
	sm.A()
	sm.B()
	sm.B()
	sm.A()
	sm.close()
}
