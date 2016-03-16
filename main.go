package main

import (
	"flag"
	"fmt"

	"github.com/SvenDowideit/ddnsclient/protocols"

	// import all the protocol drivers
	_ "github.com/SvenDowideit/ddnsclient/cloudflare"
	_ "github.com/SvenDowideit/ddnsclient/dreamhost"
	_ "github.com/SvenDowideit/ddnsclient/noip"
)

var help = flag.Bool("help", false, "Show Help")

var login = flag.String("login", "", "login to log into your account")
var password = flag.String("password", "", "password to log into your account")
var host = flag.String("host", "test.demo.gallery", "host to inspect / modify")
var ip = flag.String("ip", "10.10.10.10", "IP address to set to")

var protocol = flag.String("protocol", "",
	fmt.Sprintf("DDNS service providor\n\tone of: %s\n", protocols.ListProtocols()))

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	fmt.Printf("Set %s to %s\n", *host, *ip)
	c, err := protocols.CreateNew(*protocol, *host, *ip, *login, *password)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	c.Set()
}
