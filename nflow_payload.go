package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
	"time"
	"strings"
	"strconv"
)

// Start time for this instance, used to compute sysUptime
var StartTime = time.Now().UnixNano()

// current sysUptime in msec - recalculated in CreateNFlowHeader()
var sysUptime uint32 = 0

// Counter of flow packets that have been sent
var flowSequence uint32 = 0

const (
	UINT16_MAX      = 65535
	PAYLOAD_AVG_MD  = 1024
)

// struct data from fach
type NetflowHeader struct {
	Version        uint16
	FlowCount      uint16
	SysUptime      uint32
	UnixSec        uint32
	UnixMsec       uint32
	FlowSequence   uint32
	EngineType     uint8
	EngineId       uint8
	SampleInterval uint16
}

type NetflowPayload struct {
	SrcIP          uint32
	DstIP          uint32
	NextHopIP      uint32
	SnmpInIndex    uint16
	SnmpOutIndex   uint16
	NumPackets     uint32
	NumOctets      uint32
	SysUptimeStart uint32
	SysUptimeEnd   uint32
	SrcPort        uint16
	DstPort        uint16
	Padding1       uint8
	TcpFlags       uint8
	IpProtocol     uint8
	IpTos          uint8
	SrcAsNumber    uint16
	DstAsNumber    uint16
	SrcPrefixMask  uint8
	DstPrefixMask  uint8
	Padding2       uint16
}

//Complete netflow records
type Netflow struct {
	Header  NetflowHeader
	Records []NetflowPayload
}

//Marshall NetflowData into a buffer
func BuildNFlowPayload(data Netflow) bytes.Buffer {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, &data.Header)
	if err != nil {
		log.Println("Writing netflow header failed:", err)
	}
	for _, record := range data.Records {
		err := binary.Write(buffer, binary.BigEndian, &record)
		if err != nil {
			log.Println("Writing netflow record failed:", err)
		}
	}
	return *buffer
}

//Generate a netflow packet w/ user-defined record count
func GenerateNetflow(ipList []string) Netflow {
	data := new(Netflow)
	header := CreateNFlowHeader(len(ipList))
	records := []NetflowPayload{}
//srcIp string, destIp string, srcPort int, protocol int, subnet int
	for _, ipPair := range ipList {
		srcIpPort := strings.Split(ipPair, "-")[0]
		destIpPort := strings.Split(ipPair, "-")[1]

		srcIP := strings.Split(srcIpPort, ":")[0]
		srcPort,_ := strconv.Atoi(strings.Split(srcIpPort, ":")[1])

		destIP := strings.Split(destIpPort, ":")[0]
		destPort,_ := strconv.Atoi(strings.Split(destIpPort, ":")[1])

		records = append(records, CreateParameterizedFlow(srcIP, destIP, srcPort, destPort, 6, 24))
	}

	data.Header = header
	data.Records = records
	return *data
}

//Generate and initialize netflow header
func CreateNFlowHeader(recordCount int) NetflowHeader {

	t := time.Now().UnixNano()
	sec := t / int64(time.Second)
	nsec := t - sec*int64(time.Second)
	sysUptime = uint32((t-StartTime) / int64(time.Millisecond))+1000
	flowSequence++

	h := new(NetflowHeader)
	h.Version = 5
	h.FlowCount = uint16(recordCount)
	h.SysUptime = sysUptime
	h.UnixSec = uint32(sec)
	h.UnixMsec = uint32(nsec)
	h.FlowSequence = flowSequence
	h.EngineType = 1
	h.EngineId = 0
	h.SampleInterval = 0
	return *h
}

func CreateParameterizedFlow(srcIp string, destIp string, srcPort int, dstPort int, protocol int, subnet int) NetflowPayload {
	payload := new(NetflowPayload)
	log.Printf("here: " + srcIp + " " + destIp)
	payload.SrcIP = IPtoUint32(srcIp)
	payload.DstIP = IPtoUint32(destIp)
	payload.NextHopIP = IPtoUint32("172.199.15.1")
	payload.SrcPort = uint16(srcPort)
	payload.DstPort = uint16(dstPort)

	FillCommonFields(payload, PAYLOAD_AVG_MD, protocol, subnet)
	return *payload
}

// patch up the common fields of the packets
func FillCommonFields (
		payload *NetflowPayload, 
		numPktOct int, 
		ipProtocol int, 
		srcPrefixMask int) NetflowPayload {


	payload.NumPackets = genRandUint32(numPktOct)
	payload.NumOctets = genRandUint32(numPktOct)
	payload.Padding1 = 0
	payload.IpProtocol = uint8(ipProtocol)
	payload.IpTos = 0
	payload.SrcAsNumber = genRandUint16(UINT16_MAX)
	payload.DstAsNumber = genRandUint16(UINT16_MAX)

	payload.SrcPrefixMask = uint8(srcPrefixMask)
	payload.DstPrefixMask = uint8(rand.Intn(32))
	payload.Padding2 = 0

	// now handle computed values
	if !opts.FalseIndex {                       // default interfaces are zero
		payload.SnmpInIndex = 0
		payload.SnmpOutIndex = 0
	} else if payload.SrcIP > payload.DstIP {   // false-index
		payload.SnmpInIndex = 1
		payload.SnmpOutIndex = 2
	} else {
		payload.SnmpInIndex = 2
		payload.SnmpOutIndex = 1
	}

	uptime := int(sysUptime)
	payload.SysUptimeEnd = uint32(uptime - randomNum(10,500))
	payload.SysUptimeStart = payload.SysUptimeEnd - uint32(randomNum(10,500))

	log.Infof("S&D : %x %x %d, %d", payload.SrcIP, payload.DstIP, payload.DstPort, payload.SnmpInIndex)
	log.Infof("Time: %d %d %d", sysUptime, payload.SysUptimeStart, payload.SysUptimeEnd)

	return *payload
}

func genRandUint16(max int) uint16 {
	return uint16(rand.Intn(max))
}

func IPtoUint32(s string) uint32 {
	ip := net.ParseIP(s)
	return binary.BigEndian.Uint32(ip.To4())
}

func genRandUint32(max int) uint32 {
	return uint32(rand.Intn(max))
}
