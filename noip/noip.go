
package noip

import (
//   "encoding/base64"
    "fmt"
    "strings"
        
    "github.com/SvenDowideit/ddnsclient/protocols"
)

// See http://www.noip.com/integrate/request

// TODO: can I get the package name for this?
const driverName = "noip"

func init() {
    protocols.RegisterDriver(driverName, New)
}

func New(host, ip, login, password string) (protocols.Driver, error) {
    c := Noip{
        host:host,
        ip:ip, 
        login:login,
        password:password,
        
        apiUrl: "https://dynupdate.no-ip.com/nic",
        
        headers: map[string]string{
            "X-Auth-Email": login,
            "X-Auth-Key": password,
        },
    }
    
    return c, nil
}

type Noip struct {
    host, ip, login, password string
        
    apiUrl string
    headers map[string]string
}

//TODO: add protocol support for 'Get()'

func (c Noip) Set() {
    // http://username:password@dynupdate.no-ip.com/nic/update?hostname=mytest.testdomain.com&myip=1.2.3.4
   
   url := "https://"+c.login+":"+c.password+"@dynupdate.no-ip.com/nic/update"
   
    resultCode, err := get(url, map[string]string{
            "hostname": c.host,
            "myip": c.ip,
//            "Authorization": "Basic "+base64.StdEncoding.EncodeToString([]byte(c.login + ":" + c.password)),
        }, c.headers)
    if err != nil {
        fmt.Println("ERROR: ", err)
        return
    }
    
    ret := strings.Split(resultCode, " ")
    fmt.Printf("%s %s to %s\n", ret[0], c.host, ret[1])
}

func get(cmd string, options, headers map[string]string) (string, error) {
    body, err := protocols.Get(cmd, options, headers)

    resultCode := strings.TrimSpace(string(body))
    if err == nil {
        switch resultCode {
        case "nohost":
            err = fmt.Errorf(resultCode, "Hostname supplied does not exist under specified account, client exit and require user to enter new login credentials before performing an additional request.")
        case "badauth":
            err = fmt.Errorf(resultCode, "Invalid username password combination")
        case "badagent":
            err = fmt.Errorf(resultCode, "Client disabled. Client should exit and not perform any more updates without user intervention. ")
        case "!donator":
            err = fmt.Errorf(resultCode, "An update request was sent including a feature that is not available to that particular user such as offline options.\nContact http://www.noip.com/support/")
        case "abuse":
            err = fmt.Errorf(resultCode, "Username is blocked due to abuse. Either for not following our update specifications or disabled due to violation of the No-IP terms of service. Our terms of service can be viewed here. Client should stop sending updates.\nContact http://www.noip.com/support/")
        case "911":
            err = fmt.Errorf(resultCode, "A fatal error on our side such as a database outage. Retry the update no sooner than 30 minutes.\nContact http://www.noip.com/support/")
        }
    }
    return resultCode, err
}

