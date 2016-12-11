package msgtypes

import (
	pb "github.com/ms-xy/Holmes-Planner-Monitor/protobuf/generated-go"
	"net"
)

// -----------------------------------------------------------------------------

type StatusMessage struct {
	PID           uint64
	UUID          *UUID
	PlannerInfo   *PlannerInfo
	SystemStatus  *SystemStatus
	NetworkStatus *NetworkStatus
	PlannerStatus *PlannerStatus
	ServiceStatus *ServiceStatus
}

func (this *StatusMessage) FromPb(o *pb.StatusMessage) *StatusMessage {
	this.PID = o.PID
	this.UUID = UUID4Empty()
	if err := this.UUID.FromBytes(o.UUID); err != nil {
		panic(err) // TODO: is there a better way than to panic?
	}
	if o.PlannerInfo != nil {
		this.PlannerInfo = (&PlannerInfo{}).FromPb(o.PlannerInfo)
	} else if o.SystemStatus != nil {
		this.SystemStatus = (&SystemStatus{}).FromPb(o.SystemStatus)
	} else if o.NetworkStatus != nil {
		this.NetworkStatus = (&NetworkStatus{}).FromPb(o.NetworkStatus)
	} else if o.PlannerStatus != nil {
		this.PlannerStatus = (&PlannerStatus{}).FromPb(o.PlannerStatus)
	} else if o.ServiceStatus != nil {
		this.ServiceStatus = (&ServiceStatus{}).FromPb(o.ServiceStatus)
	}
	return this
}

func (this *StatusMessage) ToPb() *pb.StatusMessage {
	if this.UUID == nil {
		this.UUID = UUID4Empty()
	}
	o := &pb.StatusMessage{
		PID:  this.PID,
		UUID: this.UUID.ToBytes(),
	}
	if this.PlannerInfo != nil {
		o.PlannerInfo = this.PlannerInfo.ToPb()
	} else if this.SystemStatus != nil {
		o.SystemStatus = this.SystemStatus.ToPb()
	} else if this.NetworkStatus != nil {
		o.NetworkStatus = this.NetworkStatus.ToPb()
	} else if this.ServiceStatus != nil {
		o.ServiceStatus = this.ServiceStatus.ToPb()
	}
	return o
}

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

	Harddrives []*Harddrive

	Loads1  float64 // System load as reported by sysinfo syscall
	Loads5  float64
	Loads15 float64
}

func (this *SystemStatus) FromPb(o *pb.SystemStatus) *SystemStatus {
	this.Uptime = int64(o.Uptime)
	this.LoadPercent = o.LoadPercent

	this.MemoryUsage = o.MemoryUsage
	this.MemoryMax = o.MemoryMax

	this.Harddrives = make([]*Harddrive, len(o.Harddrives))
	for i := 0; i < len(o.Harddrives); i++ {
		this.Harddrives[i] = &Harddrive{
			FsType:     o.Harddrives[i].FsType,
			MountPoint: o.Harddrives[i].MountPoint,
			Used:       o.Harddrives[i].Used,
			Total:      o.Harddrives[i].Total,
			Free:       o.Harddrives[i].Free,
		}
	}

	this.Loads1 = o.Loads1
	this.Loads5 = o.Loads5
	this.Loads15 = o.Loads15
	return this
}

func (this *SystemStatus) ToPb() *pb.SystemStatus {
	o := &pb.SystemStatus{
		Uptime:      uint64(this.Uptime),
		LoadPercent: this.LoadPercent,

		MemoryUsage: this.MemoryUsage,
		MemoryMax:   this.MemoryMax,

		Loads1:  this.Loads1,
		Loads5:  this.Loads5,
		Loads15: this.Loads15,
	}
	o.Harddrives = make([]*pb.Harddrive, len(this.Harddrives))
	for i := 0; i < len(this.Harddrives); i++ {
		o.Harddrives[i] = &pb.Harddrive{
			FsType:     this.Harddrives[i].FsType,
			MountPoint: this.Harddrives[i].MountPoint,
			Used:       this.Harddrives[i].Used,
			Total:      this.Harddrives[i].Total,
			Free:       this.Harddrives[i].Free,
		}
	}
	return o
}

type NetworkStatus struct {
	Interfaces []*NetworkInterface
}

func (this *NetworkStatus) FromPb(o *pb.NetworkStatus) *NetworkStatus {
	l := len(o.Interfaces)
	this.Interfaces = make([]*NetworkInterface, l)
	for i := 0; i < l; i++ {
		this.Interfaces[i] = &NetworkInterface{
			ID:        int(o.Interfaces[i].Id),
			Name:      o.Interfaces[i].Name,
			IP:        net.IP(o.Interfaces[i].Ip),
			Netmask:   net.IPMask(o.Interfaces[i].Netmask),
			Broadcast: net.IP(o.Interfaces[i].Broadcast),
			Scope:     o.Interfaces[i].Scope,
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
			Id:        int32(this.Interfaces[i].ID),
			Name:      this.Interfaces[i].Name,
			Ip:        this.Interfaces[i].IP,
			Netmask:   this.Interfaces[i].Netmask,
			Broadcast: this.Interfaces[i].Broadcast,
			Scope:     this.Interfaces[i].Scope,
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

type Harddrive struct {
	FsType     string
	MountPoint string
	Used       uint64
	Total      uint64
	Free       uint64
}

type NetworkInterface struct {
	ID        int
	Name      string
	IP        net.IP
	Netmask   net.IPMask
	Broadcast net.IP
	Scope     string
}

type StatusKvPair struct {
	Key   string
	Value string
}

// -----------------------------------------------------------------------------

type ControlMessage struct {
	UUID          *UUID
	AckConnect    bool
	AckDisconnect bool
	ExtraData     [][]byte
}

func (this *ControlMessage) FromPb(o *pb.ControlMessage) *ControlMessage {
	this.UUID = UUID4Empty()
	this.UUID.FromBytes(o.Uuid)
	this.AckConnect = o.AckConnect
	this.AckDisconnect = o.AckDisconnect
	this.ExtraData = o.ExtraData
	return this
}

func (this *ControlMessage) ToPb() *pb.ControlMessage {
	if this.UUID == nil {
		this.UUID = UUID4Empty()
	}
	return &pb.ControlMessage{
		Uuid:          this.UUID.ToBytes(),
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
