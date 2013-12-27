/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 04.12.13
 */
package hub

import (
	"sync"
	"sort"
	"time"
	"log"
	"bytes"
	"net/http"
	"selenium-hub/session"
	"selenium-hub/proxy"
	"selenium-hub/translator"
)

type Hub struct {
	nodes              map[string]*session.Node
	activeSessions     map[string]*session.Session
	availableSessions  []*session.Session
	nodesLocker        *sync.RWMutex
	activeLocker       *sync.RWMutex
	availableLocker    *sync.RWMutex
	sessionTimeout     time.Duration
	nodeTimeout        time.Duration
}

func New() *Hub {
	var hub *Hub = new(Hub)
	hub.nodes = make(map[string]*session.Node)
	hub.activeSessions = make(map[string]*session.Session)
	hub.nodesLocker = new(sync.RWMutex)
	hub.activeLocker = new(sync.RWMutex)
	hub.availableLocker = new(sync.RWMutex)
	hub.sessionTimeout = 30*time.Second
	hub.nodeTimeout = 30*time.Second
	return hub
}

func (seleniumHub *Hub) ReserveSession(capabilities *session.Capabilities) (*session.Session, bool) {
	cs := seleniumHub.getSortedSessions(*capabilities)
	if cs.Len() == 0 {
		return nil, false
	}
	sort.Sort(cs)
	var controller *session.QueueElement = new(session.QueueElement)
	controller.Actual = make(chan bool)
	controller.Session = make(chan *session.Session)
	for _, sortedCapabilities := range cs.GetIterator() {
		sortedCapabilities.Session.Register(controller)
		log.Println(sortedCapabilities.Weight, sortedCapabilities.Session.Capabilities.BrowserName)
	}
	controller.Actual <- true
	var session *session.Session = <-controller.Session
	close(controller.Actual)
	return session, true
}

func (seleniumHub *Hub) StartSession(seleniumSession *session.Session) {
	seleniumHub.activeLocker.Lock()
	defer seleniumHub.activeLocker.Unlock()
	seleniumSession.Timer = time.AfterFunc(seleniumHub.sessionTimeout, func() {
			seleniumHub.FreeSession(seleniumSession.Id)
		})
	seleniumHub.activeSessions[seleniumSession.Id] = seleniumSession
}

func (seleniumHub *Hub) FreeSession(sessionId string) {
	log.Println("Waiting lock for free session", sessionId)
	seleniumHub.activeLocker.Lock()
	defer seleniumHub.activeLocker.Unlock()
	log.Println("Free session", sessionId)
	seleniumSession := seleniumHub.activeSessions[sessionId]
	seleniumSession.Finish()
	delete(seleniumHub.activeSessions, sessionId)
}

func (seleniumHub *Hub) getSortedSessions(capabilities session.Capabilities) *session.CapabilitiesSorter {
	seleniumHub.nodesLocker.RLock()
	defer seleniumHub.nodesLocker.RUnlock()
	cs := session.NewSorter(capabilities)
	for _, session := range seleniumHub.availableSessions {
		cs.Add(session)
	}
	return cs
}

func (seleniumHub *Hub) prestartSession(seleniumSession *session.Session) {
	log.Println("Prestart session", seleniumSession.Capabilities)
	data := translator.GetCreateSessionRequestData(seleniumSession.Capabilities)
	log.Println("GetCreateSessionRequest", string(data))
	var r *http.Request = new(http.Request)
	r.Method = "POST"
	r.RequestURI = "/wd/hub/session"
	data, status, error := proxy.ProxyRequest(seleniumSession.Node.Url, r, bytes.NewReader(data))
	log.Println(string(data))
	if error == nil && status == 200 {
		answer := translator.GetCreateSessionAnswer(data)
		if answer.Status == 0 {
			seleniumSession.Status = session.Prestarted
			seleniumSession.Id = answer.SessionID
			seleniumSession.Capabilities = &answer.Value
		}
	} else {
		log.Fatalln(error, status)
	}
}

func (seleniumHub *Hub) RegisterNode(machine *translator.Proxy) (bool) {
	var sessions []*session.Session
	seleniumNode := session.NewNode(machine.Configuration.Url, machine.Configuration.MaxSession)
	for _, capabilities := range machine.Capabilities {
		if capabilities.SeleniumProtocol == "WebDriver" {
			for ; capabilities.MaxInstances > 0; capabilities.MaxInstances-- {
				sessions = append(sessions, session.New(capabilities.Capabilities, seleniumNode))
			}
			go seleniumHub.prestartSession(sessions[len(sessions) - 1])
		}
	}
	if len(sessions) > 0 {
		seleniumHub.DeleteNode(machine.Configuration.Url)
		seleniumHub.nodesLocker.Lock()
		defer seleniumHub.nodesLocker.Unlock()
		seleniumHub.availableLocker.Lock()
		defer seleniumHub.availableLocker.Unlock()
		// @todo Убивать ноду, если она уже существует, а только потом перезаписывать
		seleniumHub.availableSessions = append(seleniumHub.availableSessions, sessions...)
		seleniumHub.nodes[machine.Configuration.Url] = seleniumNode
		response := translator.GetApiProxyResponseData(machine)
		seleniumNode.ApiProxyResponse = response
		seleniumNode.Timer = time.AfterFunc(seleniumHub.nodeTimeout, func() {
				seleniumHub.DeleteNode(machine.Configuration.Url)
			})
		return true
	}
	return false
}

func (seleniumHub *Hub) DeleteNode(nodeId string) {
	log.Println("Waiting lock for delete node", nodeId)
	seleniumHub.nodesLocker.Lock()
	defer seleniumHub.nodesLocker.Unlock()
	if seleniumNode, found := seleniumHub.nodes[nodeId]; found {
		log.Println("Node", nodeId, "was found. Delete it.")
		seleniumNode.Timer.Stop()
		delete(seleniumHub.nodes, nodeId)
		seleniumHub.availableLocker.Lock()
		defer seleniumHub.availableLocker.Unlock()
		for _, seleniumSession := range seleniumHub.availableSessions {
			if seleniumSession.Node == seleniumNode {
				seleniumSession.Exit()
				if seleniumSession.Status == session.Active {
					go func() {
						seleniumHub.activeLocker.Lock()
						defer seleniumHub.activeLocker.Unlock()
						delete(seleniumHub.activeSessions, seleniumSession.Id)
					}()
				}
			}
		}
	}
}

func (seleniumHub *Hub) GetNodeData(nodeId string) ([]byte, bool) {
	seleniumHub.nodesLocker.RLock()
	defer seleniumHub.nodesLocker.RUnlock()
	if seleniumNode, found := seleniumHub.nodes[nodeId]; found {
		seleniumNode.Timer.Reset(seleniumHub.nodeTimeout)
		return seleniumNode.ApiProxyResponse, true
	}
	return nil, false
}

func (seleniumHub *Hub) GetSessionUrl(sessionId string) (string, bool) {
	seleniumHub.activeLocker.RLock()
	defer seleniumHub.activeLocker.RUnlock()
	if seleniumSession, found := seleniumHub.activeSessions[sessionId]; found {
		seleniumSession.Timer.Reset(seleniumHub.sessionTimeout)
		return seleniumSession.Node.Url, true
	}
	return "", false
}

func (seleniumHub *Hub) GetSessions() (sessions []session.Session) {
	seleniumHub.activeLocker.RLock()
	defer seleniumHub.activeLocker.RUnlock()
	for _, session := range seleniumHub.activeSessions {
		sessions = append(sessions, *session)
	}
	return
}
