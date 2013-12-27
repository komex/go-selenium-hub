/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 26.12.13
 */
package session

import (
	"time"
)

type Capabilities struct {
	BrowserName     string   `json:"browserName"`
	Version         property `json:"version"`
	Platform        property `json:"platform"`
}

type Session struct {
	Id           string        `json:"id"`
	Capabilities *Capabilities `json:"capabilities"`
	Status       uint8         `json:"-"`
	Timer        *time.Timer   `json:"-"`
	Node         *Node         `json:"-"`
	queue        []*QueueElement
	command      chan command
	action       func (*QueueElement)
}

func (capabilities *Session) registerElement(element *QueueElement) {
	capabilities.queue = append(capabilities.queue, element)
}

func (session *Session) start(element *QueueElement) {
	if _, actual := <-element.Actual; actual {
		element.Session <- session
		close(element.Session)
		session.action = session.registerElement
		session.Status = Active
	}
}

func (session *Session) processor() {
	for command := range session.command {
		switch command.cmd {
		case register:
			element := (command.arguments).(*QueueElement)
			session.action(element)
		case exit:
			session.Timer.Stop()
			session.queue = nil
			close(session.command)
		case finish:
			session.Id = ""
			session.Timer.Stop()
			var position int
			for index, element := range session.queue {
				position = index + 1
				if _, actual := <-element.Actual; actual {
					element.Session <- session
					close(element.Session)
					session.Status = Active
					break
				}
			}
			session.Status = Available
			session.queue = session.queue[position:]
			if len(session.queue) == 0 {
				session.action = session.start
			}
		case getWeight:
			channel := (command.arguments).(chan int)
			scores := len(session.queue)
			if scores == 0 {
				switch session.Status {
				case Prestarted:
					scores = -10
				case Active:
					scores = 10
				}
			}
			channel <- scores
		}
	}
}

func (capabilities *Session) Register(element *QueueElement) {
	capabilities.command <- command{register, element}
}

func (capabilities *Session) Finish() {
	capabilities.command <- command{finish, nil}
}

func (capabilities *Session) Exit() {
	capabilities.command <- command{exit, nil}
}

func (capabilities *Session) GetWeight() int {
	weight := make(chan int)
	defer close(weight)
	capabilities.command <- command{getWeight, weight}
	return <-weight
}

func New(capabilities Capabilities, seleniumNode *Node) *Session {
	var session *Session = new(Session)
	seleniumNode.RegisterSession(session)
	session.Capabilities = &capabilities
	session.Status = Available
	session.Node = seleniumNode
	session.command = make(chan command, 10)
	session.action = session.start
	session.Timer = new(time.Timer)
	go session.processor()
	return session
}

type property string

func (property property) Any() bool {
	return property == "ANY" || property == ""
}

type QueueElement struct {
	Actual  chan bool
	Session chan *Session
}

const (
	register uint8 = iota
	finish
	getWeight
	exit
)

const (
	Prestarted uint8 = iota
	Available
	Active
)

type command struct {
	cmd       uint8
	arguments interface {}
}
