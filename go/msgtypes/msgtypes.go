package msgtypes

import (
	pb "github.com/ms-xy/Holmes-Planner-Monitor/protobuf/generated-go"
	"net"
)

// -----------------------------------------------------------------------------

type PlannerInfo struct {
	Name          string
	ListenAddress *net.TCPAddr
	Connect       bool
	Disconnect    bool
}

func (this *PlannerInfo) FromPb(o *pb.PlannerInfo) *PlannerInfo {
	this.Name = o.Name
	if o.ListenAddress != "" {
		this.ListenAddress, _ = net.ResolveTCPAddr("tcp", o.ListenAddress)
		// TODO: error handling, we just got a malformed PlannerInfo ...
	}
	this.Connect = o.Connect
	this.Disconnect = o.Disconnect
	return this
}

func (this *PlannerInfo) ToPb() *pb.PlannerInfo {
	listenAddress := ""
	if this.ListenAddress != nil {
		listenAddress = this.ListenAddress.String()
	}
	return &pb.PlannerInfo{
		Name:          this.Name,
		ListenAddress: listenAddress,
		Connect:       this.Connect,
		Disconnect:    this.Disconnect,
	}
}

// -----------------------------------------------------------------------------

type SystemStatus struct {
	Uptime      int64
	LoadPercent float64 // Average load of all cpus during the last 1 second

	MemoryUsage uint64
	MemoryMax   uint64
	DiskUsage   uint64
	DiskMax     uint64

	Loads1  float64 // System load as reported by sysinfo syscall
	Loads5  float64
	Loads15 float64
}

func (this *SystemStatus) FromPb(o *pb.SystemStatus) *SystemStatus {
	this.Uptime = int64(o.Uptime)
	this.LoadPercent = o.LoadPercent

	this.MemoryUsage = o.MemoryUsage
	this.MemoryMax = o.MemoryMax
	this.DiskUsage = o.DiskUsage
	this.DiskMax = o.DiskMax

	this.Loads1 = o.Loads1
	this.Loads5 = o.Loads5
	this.Loads15 = o.Loads15
	return this
}

func (this *SystemStatus) ToPb() *pb.SystemStatus {
	return &pb.SystemStatus{
		Uptime:      uint64(this.Uptime),
		LoadPercent: this.LoadPercent,

		MemoryUsage: this.MemoryUsage,
		MemoryMax:   this.MemoryMax,
		DiskUsage:   this.DiskUsage,
		DiskMax:     this.DiskMax,

		Loads1:  this.Loads1,
		Loads5:  this.Loads5,
		Loads15: this.Loads15,
	}
}

type NetworkStatus struct {
	Interfaces []*NetworkInterface
}

func (this *NetworkStatus) FromPb(o *pb.NetworkStatus) *NetworkStatus {
	l := len(o.Interfaces)
	this.Interfaces = make([]*NetworkInterface, l)
	for i := 0; i < l; i++ {
		this.Interfaces[i] = &NetworkInterface{
			Name:     o.Interfaces[i].Name,
			Hwaddr:   o.Interfaces[i].Hwaddr,
			Inetaddr: o.Interfaces[i].Inetaddr,
			Netmask:  o.Interfaces[i].Netmask,
			Iface:    o.Interfaces[i].Iface,
		}
	}
	return this
}

func (this *NetworkStatus) ToPb() *pb.NetworkStatus {
	l := len(this.Interfaces)
	o := &pb.NetworkStatus{}
	o.Interfaces = make([]*pb.NetworkInterface, l)
	for i := 0; i < l; i++ {
		o.Interfaces[i] = &pb.NetworkInterface{
			Name:     this.Interfaces[i].Name,
			Hwaddr:   this.Interfaces[i].Hwaddr,
			Inetaddr: this.Interfaces[i].Inetaddr,
			Netmask:  this.Interfaces[i].Netmask,
			Iface:    this.Interfaces[i].Iface,
		}
	}
	return o
}

type PlannerStatus struct {
	ConfigProfileName string
	Logs              []string
	ExtraData         [][]byte
}

func (this *PlannerStatus) FromPb(o *pb.PlannerStatus) *PlannerStatus {
	this.ConfigProfileName = o.ConfigProfileName
	this.Logs = o.Logs
	this.ExtraData = o.ExtraData
	return this
}

func (this *PlannerStatus) ToPb() *pb.PlannerStatus {
	return &pb.PlannerStatus{
		ConfigProfileName: this.ConfigProfileName,
		Logs:              this.Logs,
		ExtraData:         this.ExtraData,
	}
}

type ServiceStatus struct {
	ConfigProfileName string
	Name              string
	Port              uint16
	Task              string
	Logs              []string
	ExtraData         [][]byte
}

func (this *ServiceStatus) FromPb(o *pb.ServiceStatus) *ServiceStatus {
	this.ConfigProfileName = o.ConfigProfileName
	this.Name = o.Name
	this.Port = uint16(o.Port)
	this.Task = o.Task
	this.Logs = o.Logs
	this.ExtraData = o.ExtraData
	return this
}

func (this *ServiceStatus) ToPb() *pb.ServiceStatus {
	return &pb.ServiceStatus{
		ConfigProfileName: this.ConfigProfileName,
		Name:              this.Name,
		Port:              uint32(this.Port),
		Task:              this.Task,
		Logs:              this.Logs,
		ExtraData:         this.ExtraData,
	}
}

// -----------------------------------------------------------------------------

type NetworkInterface struct {
	Name     string
	Hwaddr   string
	Inetaddr string
	Netmask  string
	Iface    string
}

type StatusKvPair struct {
	Key   string
	Value string
}

// -----------------------------------------------------------------------------

type ControlMessage struct {
	AckConnect    bool
	AckDisconnect bool
	ExtraData     [][]byte
}

func (this *ControlMessage) FromPb(o *pb.ControlMessage) *ControlMessage {
	this.AckConnect = o.AckConnect
	this.AckDisconnect = o.AckDisconnect
	this.ExtraData = o.ExtraData
	return this
}

func (this *ControlMessage) ToPb() *pb.ControlMessage {
	return &pb.ControlMessage{
		AckConnect:    this.AckConnect,
		AckDisconnect: this.AckDisconnect,
		ExtraData:     this.ExtraData,
	}
}

// type Message interface {
// 	String() string
// 	Encode() []byte
// 	Decode([]byte)
// }

// /*
// Possible method 1)
//     Have one Monitor object - Singleton - that hands out channels for
//     connection.
//     Loops over all channels and checks if a channel has data to send

// Possible method 2)
//     Have a factory return a reference to a Monitor (slave) that corresponds
//     via a master (has the advantage of wrapping)
// */

// func NewStringMessage(s string) *StringMessage {
// 	msg := StringMessage(s)
// 	return &msg
// }

// type StringMessage string

// func (s *StringMessage) String() string {
// 	return string(*s)
// }
// func (s *StringMessage) Encode() []byte {
// 	return []byte(string(*s))
// }
// func (s *StringMessage) Decode(x []byte) {
// 	*s = StringMessage(string(x))
// }

// func NewIntMessage(i int) *IntMessage {
// 	msg := IntMessage(i)
// 	return &msg
// }

// type IntMessage int

// func (i *IntMessage) String() string {
// 	return strconv.FormatInt(int64(int(*i)), 10)
// }
// func (i *IntMessage) Encode() []byte {
// 	return []byte(i.String())
// }
// func (i *IntMessage) Decode(x []byte) {
// 	i32, _ := strconv.ParseInt(string(x), 10, 32)
// 	*i = IntMessage(int(i32))
// }

func kvPairsToPb(inPairs []*StatusKvPair) []*pb.StatusKvPair {
	l := len(inPairs)
	outPairs := make([]*pb.StatusKvPair, l)
	for i := 0; i < l; i++ {
		outPairs[i] = &pb.StatusKvPair{
			inPairs[i].Key,
			inPairs[i].Value,
		}
	}
	return outPairs
}

func kvPairsFromPb(inPairs []*pb.StatusKvPair) []*StatusKvPair {
	l := len(inPairs)
	outPairs := make([]*StatusKvPair, l)
	for i := 0; i < l; i++ {
		outPairs[i] = &StatusKvPair{
			inPairs[i].Key,
			inPairs[i].Value,
		}
	}
	return outPairs
}
