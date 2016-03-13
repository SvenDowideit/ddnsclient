
package protocols

import (
    "fmt"
)

func RegisterDriver(name string, creator NewFunc) {
    fmt.Printf("Registered %s\n", name)
    driverRegistry[name] = creator
}
func CreateNew(name, host, ip, login, password string, verbose, debug bool) (Driver, error) {
    newdriver, ok := driverRegistry[name]
    if !ok {
        return nil, fmt.Errorf("No protocol driver found for '%s'", name)
    }
    return newdriver(host, ip, login, password, verbose, debug)    
}

type Driver interface {
    Set()
}

type NewFunc func(host, ip, login, password string, verbose, debug bool) (Driver, error)

var driverRegistry = make(map[string]NewFunc)

