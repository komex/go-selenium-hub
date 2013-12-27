/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 26.12.13
 */
package session

import (
	"time"
)

type Node struct {
	Url              string
	ApiProxyResponse []byte
	maxSessions      uint8
	Timer            *time.Timer
	sessions         []*Session
}

func (node *Node) RegisterSession(session *Session) {
	node.sessions = append(node.sessions, session)
}

func NewNode(url string, maxSessions uint8) *Node {
	var node *Node = new(Node)
	if maxSessions <= 0 {
		maxSessions = 5
	}
	node.Url = url
	node.maxSessions = maxSessions
	return node
}
