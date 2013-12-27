/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 04.12.13
 */
package main

import (
	"time"
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
	"bytes"
	"runtime"
	"log"
	"selenium-hub/proxy"
	"selenium-hub/translator"
	"selenium-hub/session"
	"selenium-hub/hub"
)

type answer struct {
	Message          string `json:"message"`
	LocalizedMessage string `json:"localizedMessage"`
}

var seleniumHub = hub.New()

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/grid/register", httpRegisterProxy).Methods("POST")
	router.HandleFunc("/grid/api/proxy", httpApiProxy).Methods("GET").Queries("id", "")
	router.HandleFunc("/wd/hub/status", httpStatus).Methods("GET")
	router.HandleFunc("/wd/hub/sessions", httpGetSessions).Methods("GET")
	router.HandleFunc("/wd/hub/session", httpCreateSession).Methods("POST")
	router.HandleFunc("/wd/hub/session/{session:[a-f0-9-]+}", httpFreeSession).Methods("DELETE")
	sessionRouter := router.PathPrefix("/wd/hub/session/{session:[a-f0-9-]+}").Subrouter()

	registerElementRoutes(sessionRouter)
	registerWindowRoutes(sessionRouter)
	registerTouchRoutes(sessionRouter)
	registerSessionRoutes(sessionRouter)

	http.Handle("/", router)
	server := &http.Server{
		Addr:           ":4444",
		Handler:        nil,
		ReadTimeout:    15*time.Minute,
		WriteTimeout:   15*time.Minute,
		MaxHeaderBytes: 1<<20,
	}
	server.ListenAndServe()
}

func httpCreateSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	log.Println("Received a create new sessions request")
	var buffer bytes.Buffer
	buffer.ReadFrom(r.Body)
	capabilities, err := translator.GetCreateSessionCapabilities(buffer.Bytes())
	if err != nil {
		http.Error(w, "Invalid capabilities.", http.StatusMethodNotAllowed)
		return
	}
	if seleniumSession, reserved := seleniumHub.ReserveSession(capabilities); reserved {
		if seleniumSession.Status == session.Prestarted {
			seleniumHub.StartSession(seleniumSession)
			setHttpHeaders(w)
			w.Write(translator.GetCreateSessionAnswerData(seleniumSession))
			return
		} else {
			data, status, error := proxy.ProxyRequest(seleniumSession.Node.Url, r, bytes.NewReader(buffer.Bytes()))
			if error != nil {
				answer := answer{}
				answer.Message = error.Error()
				answer.LocalizedMessage = error.Error()
				response(w, 13, answer)
			} else if status == 200 {
				seleniumSessionAnswer := translator.GetCreateSessionAnswer(data)
				if seleniumSessionAnswer.Status == 0 {
					seleniumSession.Id = seleniumSessionAnswer.SessionID
					seleniumHub.StartSession(seleniumSession)
					setHttpHeaders(w)
					w.Write(data)
					return
				}
			}
		}
		seleniumSession.Finish()
	} else {
		answer := answer{}
		answer.Message = "Session for required capabilities was not found."
		answer.LocalizedMessage = "Сессия, подходящая под запрашиваемые требования не найдена."
		response(w, 13, answer)
	}
}

func httpFreeSession(w http.ResponseWriter, r *http.Request) {
	sessionId := mux.Vars(r)["session"]
	proxySessionRequest(w, r)
	seleniumHub.FreeSession(sessionId)
}

func httpRegisterProxy(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a register new selenium node request")
	setHttpHeaders(w)
	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	machine, _ := translator.GetProxy(r.Body)
	if registered := seleniumHub.RegisterNode(machine); registered {
		w.Write([]byte("ok"))
	} else {
		answer := answer{}
		answer.Message = "Node does not have WebDriver sessions."
		answer.LocalizedMessage = "seleniumNode не поддерживает работу с WebDriver."
		response(w, 13, answer)
	}
}

func httpGetSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	sessions := seleniumHub.GetSessions()
	if len(sessions) == 0 {
		response(w, 0, []session.Session{})
	} else {
		response(w, 0, sessions)
	}
}

func httpApiProxy(w http.ResponseWriter, r *http.Request) {
	setHttpHeaders(w)
	var nodeId string = r.FormValue("id")
	if data, found := seleniumHub.GetNodeData(nodeId); found {
		w.Write(data)
	} else {
		http.NotFound(w, r)
	}
}

func httpStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("DD")
	type os struct {
		Name string `json:"name"`
		Arch string `json:"arch"`
	}
	type value struct {
		OS os `json:"os"`
	}
	response(w, 0, value{os{runtime.GOOS, runtime.GOARCH}})
}

func response(w http.ResponseWriter, status uint8, value interface {}) {
	data := translator.GetResponse(status, value)
	setHttpHeaders(w)
	w.Write(data)
}

func proxySessionRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Proxy session request:", r.Method, r.URL)
	sessionId := mux.Vars(r)["session"]
	setHttpHeaders(w)
	if url, found := seleniumHub.GetSessionUrl(sessionId); found {
		data, status, error := proxy.ProxyRequest(url, r, r.Body)
		if error != nil {
			log.Fatalln("Error while proxy request: %s\n", error)
			answer := answer{}
			answer.Message = error.Error()
			answer.LocalizedMessage = error.Error()
			response(w, 13, answer)
			return
		}
		w.WriteHeader(status)
		w.Write(data)
	} else {
		log.Println("Session", sessionId, "not found")
		answer := answer{}
		answer.Message = fmt.Sprintf("Session %s not found.", sessionId)
		answer.LocalizedMessage = fmt.Sprintf("Сессия %s не найдена.", sessionId)
		response(w, 13, answer)
	}
}

func setHttpHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Server", "go-selenium-hub")
}
