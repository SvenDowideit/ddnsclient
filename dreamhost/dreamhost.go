
package noip

import (
//   "encoding/base64"
    "fmt"
    "crypto/rand"
    "strings"
        
    "github.com/SvenDowideit/ddnsclient/protocols"
)

// See http://wiki.dreamhost.com/Application_programming_interface

// TODO: can I get the package name for this?
const driverName = "dreamhost"

func init() {
    protocols.RegisterDriver(driverName, New)
}

func New(host, ip, login, password string) (protocols.Driver, error) {
    c := Dreamhost{
        host:host,
        ip:ip, 
        login:login,
        password:password,
        
        apiUrl: "https://api.dreamhost.com",
        
        headers: map[string]string{
            "X-Auth-Email": login,
            "X-Auth-Key": password,
        },
    }
    
    return c, nil
}

type Dreamhost struct {
    host, ip, login, password string
        
    apiUrl string
    headers map[string]string
}

func pseudo_uuid() (uuid string) {

    b := make([]byte, 16)
    _, err := rand.Read(b)
    if err != nil {
        fmt.Println("Error: ", err)
        return
    }

    uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

    return
}

//TODO: add protocol support for 'Get()'

func (c Dreamhost) Set() {
    //https://api.dreamhost.com/?key=6SHU5P2HLDAYECUM&cmd=user-list_users_no_pw&unique_id=4082432&format=perl

//TODO: if there already is a record, it seems the only way to change the ip, is to remove it and go again.
    IP, err := c.GetIP(c.host)
    if err != nil {
        fmt.Println("ERROR: ", err)
        return
    }   
   
    fmt.Println("-----------------------------")

    // Dreamhost's API has no update, you remove and re-add
    if IP != "" {
        fmt.Println(c.host, "already exists at ", IP, "removing record first")
        resultCode, err := get(c.apiUrl, map[string]string{
                "key": c.password,
                "cmd": " dns-remove_record",
                "unique_id": pseudo_uuid(),
                "format": "json",
                "record": c.host,
                "type": "A",
                "value": IP,
            }, c.headers)
        if err != nil {
            fmt.Println("ERROR: ", err)
            return
        }
        fmt.Println(resultCode)
    }
    fmt.Println("Adding", c.host, "at", c.ip)
    resultCode, err := get(c.apiUrl, map[string]string{
            "key": c.password,
            "cmd": "dns-add_record",
            "unique_id": pseudo_uuid(),
            "format": "json",
            "record": c.host,
            "type": "A",
            "value": c.ip,
            "comment": "set by ddnsclient",
        }, c.headers)
    if err != nil {
        fmt.Println("ERROR: ", err)
        return
    }
    
    fmt.Printf("Get returned %s\n", resultCode)
}

func (c Dreamhost) List() string {
    resultCode, err := get(c.apiUrl, map[string]string{
            "key": c.password,
            "cmd": "dns-list_records",
            "unique_id": pseudo_uuid(),
        }, c.headers)
    if err != nil {
        fmt.Println("ERROR: ", err)
        return resultCode
    }
    
    //fmt.Printf("List returned %s\n", resultCode)
    return resultCode
}

func (c Dreamhost) GetIP(host string) (string, error) {
    allIPs := c.List()
    for _, line := range strings.Split(allIPs, "\n") {
        // 163611#fi.gy#ddnsclient.fi.gy#A#66.66.66.88#set#by#ddnsclient#1
        values := strings.Fields(line)
        if len(values) > 5 && values[2] == c.host && values[3] == "A" {
            return values[4], nil
        }
    }
    return "", nil
}

func get(cmd string, options, headers map[string]string) (string, error) {
    body, err := protocols.Get(cmd, options, headers)

    resultCode := strings.TrimSpace(string(body))

    return resultCode, err
}

