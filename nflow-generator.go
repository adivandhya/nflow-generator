// Run using:
// go run nflow-generator.go nflow_logging.go nflow_payload.go  -t 172.16.86.138 -p 9995
// Or:
// go build
// ./nflow-generator -t <ip> -p <port>
package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"math/rand"
	"net"
	"os"
	"strings"
)

var opts struct {
	CollectorIP   string `short:"t" long:"target" description:"target ip address of the netflow collector"`
	CollectorPort string `short:"p" long:"port" description:"port number of the target netflow collector"`
	SpikeProto    string `short:"s" long:"spike" description:"run a second thread generating a spike for the specified protocol"`
    FalseIndex    bool   `short:"f" long:"false-index" description:"generate false SNMP interface indexes, otherwise set to 0"`
    Help          bool   `short:"h" long:"help" description:"show nflow-generator help"`
	Protocol      uint	`short:"l" long:"protocol" description:"netflow packet protocol"`
    Time		  uint	`short:"e" long:"time" description:"netflow packet event time"`
    IpRange       string `short:"i" long:"iprange" description:"specify the list of ip address combinations for the netflow. eg: 172.21.1.1:8023-192.3.4.3:80,10.1.2.3:34565-4.3.2.5:443"`
}

func main() {
	_, err := flags.Parse(&opts)
	fmt.Printf("\n%+v\n", opts)
	if err != nil {
		showUsage()
		os.Exit(1)
	}
	if opts.Help == true {
		showUsage()
		os.Exit(1)
	}
	if opts.CollectorIP == "" || opts.CollectorPort == "" {

		showUsage()
		os.Exit(1)
	}
	collector := opts.CollectorIP + ":" + opts.CollectorPort
	udpAddr, err := net.ResolveUDPAddr("udp", collector)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("Error connecting to the target collector: ", err)
	}
	log.Infof("sending netflow data to a collector ip: %s and port: %s. \n"+
		"Use ctrl^c to terminate the app.", opts.CollectorIP, opts.CollectorPort)

	ipRange := opts.IpRange
	fmt.Printf(ipRange)
	ipList := strings.Split(ipRange, ",")
	data := GenerateNetflow(ipList, uint32(opts.Time), int(opts.Protocol))
	buffer := BuildNFlowPayload(data)
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Fatal("Error connecting to the target collector: ", err)
	}
}

func randomNum(min, max int) int {
	return rand.Intn(max-min) + min
}

func showUsage() {
	var usage string
	usage = `
Usage:
  main [OPTIONS] [collector IP address] [collector port number]

  Send mock Netflow version 5 data to designated collector IP & port.
  Time stamps in all datagrams are set to UTC.

Application Options:
  -t, --target= target ip address of the netflow collector
  -p, --port=   port number of the target netflow collector
  -s, --spike run a second thread generating a spike for the specified protocol
    protocol options are as follows:
        ftp - generates tcp/21
        ssh  - generates tcp/22
        dns - generates udp/54
        http - generates tcp/80
        https - generates tcp/443
        ntp - generates udp/123
        snmp - generates ufp/161
        imaps - generates tcp/993
        mysql - generates tcp/3306
        https_alt - generates tcp/8080
        p2p - generates udp/6681
        bittorrent - generates udp/6682
  -f, --false-index generate false snmp index values of 1 or 2: If the source address > dest address, input interface is set to 1, and set to 2 otherwise,
and the output interface is set to the opposite value. Default in and out interface is 0. (Optional)

Example Usage:

    -first build from source (one time)
    go build   

    -generate default flows to device 172.16.86.138, port 9995
    ./nflow-generator -t 172.16.86.138 -p 9995 

    -generate default flows along with a spike in the specified protocol:
    ./nflow-generator -t 172.16.86.138 -p 9995 -s ssh

    -generate default flows with "false index" settings for snmp interfaces 
    ./nflow-generator -t 172.16.86.138 -p 9995 -f

Help Options:
  -h, --help    Show this help message
  `
	fmt.Print(usage)
}
