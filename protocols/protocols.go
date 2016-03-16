package protocols

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"net/http"
)

const userAgent = "https://github.com/SvenDowideit/ddnsclient DEV SvenDowideit@home.org.au"

var verbose = false
var debug = false

func init() {
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Debug output")
}

func RegisterDriver(name string, creator NewFunc) {
	if verbose {
		fmt.Printf("Registered %s\n", name)
	}
	driverRegistry[name] = creator
}
func CreateNew(name, host, ip, login, password string) (Driver, error) {
	newdriver, ok := driverRegistry[name]
	if !ok {
		return nil, fmt.Errorf("No protocol driver found for '%s'", name)
	}
	return newdriver(host, ip, login, password)
}

type Driver interface {
	Set()
}

type NewFunc func(host, ip, login, password string) (Driver, error)

var driverRegistry = make(map[string]NewFunc)

func ListProtocols() string {
	keys := make([]string, len(driverRegistry))

	i := 0
	for k := range driverRegistry {
		keys[i] = k
		i++
	}
	return strings.Join(keys, ", ")
}

func CallJSON(X, cmd string, options, headers map[string]string) ([]byte, error) {
	if verbose {
		fmt.Println(X, ": ", cmd)
	}
	client := &http.Client{}

	var buffer *bytes.Buffer
	if options != nil {
		buffer = new(bytes.Buffer)
		err := json.NewEncoder(buffer).Encode(options)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(X, cmd, buffer)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)

	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	//    req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if debug {
		var out bytes.Buffer
		json.Indent(&out, body, "", "  ")
		out.WriteTo(os.Stdout)
	}

	if verbose {
		fmt.Printf("output: %+v\n------\n", string(body))
	}

	return body, err
}

func Get(cmd string, options, headers map[string]string) ([]byte, error) {
	u, err := url.Parse(cmd)
	if verbose {
		fmt.Println("Get: ", u.String())
	}
	client := &http.Client{}

	if options != nil {
		params := url.Values{}
		for k, v := range options {
			params.Add(k, v)
		}
		u.RawQuery = params.Encode()
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)

	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if debug {
		var out bytes.Buffer
		json.Indent(&out, body, "", "  ")
		out.WriteTo(os.Stdout)
	}

	if verbose {
		fmt.Printf("output: %+v\n------\n", string(body))
	}

	return body, err
}
