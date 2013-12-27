/**
 * Author: Andrey Kolchenko <andrey@kolchenko.me>
 * Date: 26.12.13
 */
package proxy

import (
	"net/http"
	"io"
	"net"
	"io/ioutil"
	"time"
)

const browserTimeout = 30*time.Second

func ProxyRequest(url string, r *http.Request, body io.Reader) (data []byte, status int, err error) {
	var client *http.Client = new(http.Client)
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, browserTimeout)
		},
	}
	client.Transport = transport
	request, err := http.NewRequest(r.Method, url + r.RequestURI, body)
	if err != nil {
		return
	}
	response, err := client.Do(request)
	if err != nil {
		return
	}
	status = response.StatusCode
	data, err = ioutil.ReadAll(response.Body)
	return
}
