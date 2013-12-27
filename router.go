/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 19.11.13
 */
package main

import (
	"github.com/gorilla/mux"
)

func registerRoutes(router *mux.Router, api map[string][]string) {
	for path, methods := range api {
		router.HandleFunc(path, proxySessionRequest).Methods(methods...)
	}
}

func registerElementRoutes(router *mux.Router) {
	var api = make(map[string][]string)
	api["/element/active"] = []string{"POST"}
	api["/element/{id}"] = []string{"GET"}
	api["/element/{id}/element"] = []string{"POST"}
	api["/element/{id}/elements"] = []string{"POST"}
	api["/element/{id}/click"] = []string{"POST"}
	api["/element/{id}/submit"] = []string{"POST"}
	api["/element/{id}/text"] = []string{"GET"}
	api["/element/{id}/value"] = []string{"POST"}
	api["/element/{id}/name"] = []string{"GET"}
	api["/element/{id}/clear"] = []string{"POST"}
	api["/element/{id}/selected"] = []string{"GET"}
	api["/element/{id}/enabled"] = []string{"GET"}
	api["/element/{id}/attribute/{name}"] = []string{"GET"}
	api["/element/{id}/equals/{other}"] = []string{"GET"}
	api["/element/{id}/displayed"] = []string{"GET"}
	api["/element/{id}/location"] = []string{"GET"}
	api["/element/{id}/location_in_view"] = []string{"GET"}
	api["/element/{id}/size"] = []string{"GET"}
	api["/element/{id}/css/{propertyName}"] = []string{"GET"}
	registerRoutes(router, api)
}

func registerWindowRoutes(router *mux.Router) {
	var api = make(map[string][]string)
	api["/window/{windowHandle}/size"] = []string{"GET", "POST"}
	api["/window/{windowHandle}/position"] = []string{"GET", "POST"}
	api["/window/{windowHandle}/maximize"] = []string{"POST"}
	registerRoutes(router, api)
}

func registerTouchRoutes(router *mux.Router) {
	var api = make(map[string][]string)
	api["/touch/click"] = []string{"POST"}
	api["/touch/down"] = []string{"POST"}
	api["/touch/up"] = []string{"POST"}
	api["/touch/move"] = []string{"POST"}
	api["/touch/scroll"] = []string{"POST"}
	api["/touch/doubleclick"] = []string{"POST"}
	api["/touch/longclick"] = []string{"POST"}
	api["/touch/flick"] = []string{"POST"}
	registerRoutes(router, api)
}

func registerSessionRoutes(router *mux.Router) {
	var api = make(map[string][]string)
	api["/timeouts"] = []string{"POST"}
	api["/async_script"] = []string{"POST"}
	api["/implicit_wait"] = []string{"POST"}
	api["/window_handle"] = []string{"GET"}
	api["/window_handles"] = []string{"GET"}
	api["/url"] = []string{"GET", "POST"}
	api["/forward"] = []string{"POST"}
	api["/back"] = []string{"POST"}
	api["/refresh"] = []string{"POST"}
	api["/execute"] = []string{"POST"}
	api["/execute_async"] = []string{"POST"}
	api["/screenshot"] = []string{"GET"}
	api["/available_engines"] = []string{"GET"}
	api["/active_engine"] = []string{"GET"}
	api["/activated"] = []string{"GET"}
	api["/deactivate"] = []string{"POST"}
	api["/activate"] = []string{"POST"}
	api["/frame"] = []string{"POST"}
	api["/window"] = []string{"POST", "DELETE"}
	api["/cookie"] = []string{"GET", "POST", "DELETE"}
	api["/cookie/{name}"] = []string{"DELETE"}
	api["/source"] = []string{"GET"}
	api["/title"] = []string{"GET"}
	api["/element"] = []string{"POST"}
	api["/elements"] = []string{"POST"}
	api["/keys"] = []string{"POST"}
	api["/orientation"] = []string{"GET", "POST"}
	api["/alert_text"] = []string{"GET", "POST"}
	api["/accept_alert"] = []string{"POST"}
	api["/dismiss_alert"] = []string{"POST"}
	api["/moveto"] = []string{"POST"}
	api["/click"] = []string{"POST"}
	api["/buttondown"] = []string{"POST"}
	api["/buttonup"] = []string{"POST"}
	api["/doubleclick"] = []string{"POST"}
	api["/location"] = []string{"GET", "POST"}
	api["/local_storage"] = []string{"GET", "POST", "DELETE"}
	api["/local_storage/key/{key}"] = []string{"GET", "DELETE"}
	api["/local_storage/size"] = []string{"GET"}
	api["/session_storage"] = []string{"GET", "POST", "DELETE"}
	api["/session_storage/key/{key}"] = []string{"GET", "DELETE"}
	api["/session_storage/size"] = []string{"GET"}
	api["/log"] = []string{"POST"}
	api["/log/types"] = []string{"GET"}
	api["/application_cache/status"] = []string{"GET"}
	registerRoutes(router, api)
}
