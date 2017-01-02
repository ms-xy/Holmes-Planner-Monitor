// Code generated by protoc-gen-go.
// source: messages.proto
// DO NOT EDIT!

/*
Package statusMessagesProtobuf is a generated protocol buffer package.

It is generated from these files:
	messages.proto

It has these top-level messages:
	StatusMessage
	PlannerInfo
	SystemStatus
	NetworkStatus
	PlannerStatus
	ServiceStatus
	Harddrive
	NetworkInterface
	StatusKvPair
	ControlMessage
*/
package statusMessagesProtobuf

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type StatusMessage struct {
	// the pid is used for planner identification, it is not reliable in case of
	// planner restarts (or even OS restart), but reliably persistent across
	// disconnects without any restarts
	// it isn't even unique
	Pid uint64 `protobuf:"varint,5,opt,name=pid" json:"pid,omitempty"`
	// the UUID is used for machine identification, it is supposed to be
	// persistent across reboots, thus should be saved in a persistent location
	Uuid []byte `protobuf:"bytes,6,opt,name=uuid,proto3" json:"uuid,omitempty"`
	// the MachineUUID on the other hand is supposed to be persistent and thus
	// should be saved in a location that is persistent across reboots
	MachineUuid []byte `protobuf:"bytes,7,opt,name=machineUuid,proto3" json:"machineUuid,omitempty"`
	// the time stamp is important as it allows to pinpoint events to local system
	// time
	Timestamp uint64 `protobuf:"varint,8,opt,name=timestamp" json:"timestamp,omitempty"`
	// for the initial message, high number, we only send this upon initialization
	// of the connection (identifiers 1-15 use 1 byte, 16-2047 2 bytes)
	PlannerInfo *PlannerInfo `protobuf:"bytes,2048,opt,name=plannerInfo" json:"plannerInfo,omitempty"`
	// any subsequent message should contain only one of the following:
	SystemStatus  *SystemStatus  `protobuf:"bytes,1,opt,name=systemStatus" json:"systemStatus,omitempty"`
	NetworkStatus *NetworkStatus `protobuf:"bytes,2,opt,name=networkStatus" json:"networkStatus,omitempty"`
	PlannerStatus *PlannerStatus `protobuf:"bytes,3,opt,name=plannerStatus" json:"plannerStatus,omitempty"`
	ServiceStatus *ServiceStatus `protobuf:"bytes,4,opt,name=serviceStatus" json:"serviceStatus,omitempty"`
}

func (m *StatusMessage) Reset()                    { *m = StatusMessage{} }
func (m *StatusMessage) String() string            { return proto.CompactTextString(m) }
func (*StatusMessage) ProtoMessage()               {}
func (*StatusMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *StatusMessage) GetPlannerInfo() *PlannerInfo {
	if m != nil {
		return m.PlannerInfo
	}
	return nil
}

func (m *StatusMessage) GetSystemStatus() *SystemStatus {
	if m != nil {
		return m.SystemStatus
	}
	return nil
}

func (m *StatusMessage) GetNetworkStatus() *NetworkStatus {
	if m != nil {
		return m.NetworkStatus
	}
	return nil
}

func (m *StatusMessage) GetPlannerStatus() *PlannerStatus {
	if m != nil {
		return m.PlannerStatus
	}
	return nil
}

func (m *StatusMessage) GetServiceStatus() *ServiceStatus {
	if m != nil {
		return m.ServiceStatus
	}
	return nil
}

type PlannerInfo struct {
	// name is the planner's name, e.g. Holmes-Totem / Holmes-Storage / etc
	// ipAddress is the interface that the planner is listening on
	// port is the port that the planner is listening on
	Name          string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	ListenAddress string `protobuf:"bytes,3,opt,name=listenAddress" json:"listenAddress,omitempty"`
	// if a client choses to disconnect, the server should not mistake this as
	// an error but rather remove the client from its client cachex
	// additionally the server does not need to respond unless the client requests
	// a connection confirmation
	Disconnect bool `protobuf:"varint,2048,opt,name=disconnect" json:"disconnect,omitempty"`
	Connect    bool `protobuf:"varint,2049,opt,name=connect" json:"connect,omitempty"`
}

func (m *PlannerInfo) Reset()                    { *m = PlannerInfo{} }
func (m *PlannerInfo) String() string            { return proto.CompactTextString(m) }
func (*PlannerInfo) ProtoMessage()               {}
func (*PlannerInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type SystemStatus struct {
	Uptime      uint64       `protobuf:"varint,1,opt,name=uptime" json:"uptime,omitempty"`
	CpuIOWait   uint64       `protobuf:"varint,2,opt,name=cpuIOWait" json:"cpuIOWait,omitempty"`
	CpuIdle     uint64       `protobuf:"varint,3,opt,name=cpuIdle" json:"cpuIdle,omitempty"`
	CpuBusy     uint64       `protobuf:"varint,4,opt,name=cpuBusy" json:"cpuBusy,omitempty"`
	CpuTotal    uint64       `protobuf:"varint,5,opt,name=cpuTotal" json:"cpuTotal,omitempty"`
	MemoryUsage uint64       `protobuf:"varint,6,opt,name=memoryUsage" json:"memoryUsage,omitempty"`
	MemoryMax   uint64       `protobuf:"varint,7,opt,name=memoryMax" json:"memoryMax,omitempty"`
	SwapUsage   uint64       `protobuf:"varint,8,opt,name=swapUsage" json:"swapUsage,omitempty"`
	SwapMax     uint64       `protobuf:"varint,9,opt,name=swapMax" json:"swapMax,omitempty"`
	Harddrives  []*Harddrive `protobuf:"bytes,10,rep,name=harddrives" json:"harddrives,omitempty"`
	Loads1      float64      `protobuf:"fixed64,11,opt,name=Loads1,json=loads1" json:"Loads1,omitempty"`
	Loads5      float64      `protobuf:"fixed64,12,opt,name=Loads5,json=loads5" json:"Loads5,omitempty"`
	Loads15     float64      `protobuf:"fixed64,13,opt,name=Loads15,json=loads15" json:"Loads15,omitempty"`
}

func (m *SystemStatus) Reset()                    { *m = SystemStatus{} }
func (m *SystemStatus) String() string            { return proto.CompactTextString(m) }
func (*SystemStatus) ProtoMessage()               {}
func (*SystemStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SystemStatus) GetHarddrives() []*Harddrive {
	if m != nil {
		return m.Harddrives
	}
	return nil
}

type NetworkStatus struct {
	Interfaces []*NetworkInterface `protobuf:"bytes,1,rep,name=interfaces" json:"interfaces,omitempty"`
}

func (m *NetworkStatus) Reset()                    { *m = NetworkStatus{} }
func (m *NetworkStatus) String() string            { return proto.CompactTextString(m) }
func (*NetworkStatus) ProtoMessage()               {}
func (*NetworkStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *NetworkStatus) GetInterfaces() []*NetworkInterface {
	if m != nil {
		return m.Interfaces
	}
	return nil
}

type PlannerStatus struct {
	ConfigProfileName string   `protobuf:"bytes,1,opt,name=configProfileName" json:"configProfileName,omitempty"`
	Logs              []string `protobuf:"bytes,2,rep,name=logs" json:"logs,omitempty"`
	ExtraData         [][]byte `protobuf:"bytes,16,rep,name=extraData,proto3" json:"extraData,omitempty"`
}

func (m *PlannerStatus) Reset()                    { *m = PlannerStatus{} }
func (m *PlannerStatus) String() string            { return proto.CompactTextString(m) }
func (*PlannerStatus) ProtoMessage()               {}
func (*PlannerStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ServiceStatus struct {
	ConfigProfileName string   `protobuf:"bytes,1,opt,name=configProfileName" json:"configProfileName,omitempty"`
	Name              string   `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Port              uint32   `protobuf:"varint,3,opt,name=port" json:"port,omitempty"`
	Task              string   `protobuf:"bytes,4,opt,name=task" json:"task,omitempty"`
	Logs              []string `protobuf:"bytes,5,rep,name=logs" json:"logs,omitempty"`
	ExtraData         [][]byte `protobuf:"bytes,16,rep,name=extraData,proto3" json:"extraData,omitempty"`
}

func (m *ServiceStatus) Reset()                    { *m = ServiceStatus{} }
func (m *ServiceStatus) String() string            { return proto.CompactTextString(m) }
func (*ServiceStatus) ProtoMessage()               {}
func (*ServiceStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type Harddrive struct {
	FsType     string `protobuf:"bytes,1,opt,name=fsType" json:"fsType,omitempty"`
	MountPoint string `protobuf:"bytes,2,opt,name=mountPoint" json:"mountPoint,omitempty"`
	Used       uint64 `protobuf:"varint,3,opt,name=used" json:"used,omitempty"`
	Total      uint64 `protobuf:"varint,4,opt,name=total" json:"total,omitempty"`
	Free       uint64 `protobuf:"varint,5,opt,name=free" json:"free,omitempty"`
}

func (m *Harddrive) Reset()                    { *m = Harddrive{} }
func (m *Harddrive) String() string            { return proto.CompactTextString(m) }
func (*Harddrive) ProtoMessage()               {}
func (*Harddrive) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type NetworkInterface struct {
	Id        int32  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Ip        []byte `protobuf:"bytes,3,opt,name=ip,proto3" json:"ip,omitempty"`
	Netmask   []byte `protobuf:"bytes,4,opt,name=netmask,proto3" json:"netmask,omitempty"`
	Broadcast []byte `protobuf:"bytes,5,opt,name=broadcast,proto3" json:"broadcast,omitempty"`
	Scope     string `protobuf:"bytes,6,opt,name=scope" json:"scope,omitempty"`
}

func (m *NetworkInterface) Reset()                    { *m = NetworkInterface{} }
func (m *NetworkInterface) String() string            { return proto.CompactTextString(m) }
func (*NetworkInterface) ProtoMessage()               {}
func (*NetworkInterface) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type StatusKvPair struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (m *StatusKvPair) Reset()                    { *m = StatusKvPair{} }
func (m *StatusKvPair) String() string            { return proto.CompactTextString(m) }
func (*StatusKvPair) ProtoMessage()               {}
func (*StatusKvPair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type ControlMessage struct {
	// The UUID payload is important for client (re-)identification.
	Uuid        []byte `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	MachineUuid []byte `protobuf:"bytes,2,opt,name=machineUuid,proto3" json:"machineUuid,omitempty"`
	// Ack type responses are only for rarely sent messages (like planner info
	// which is only sent once at the start of a connection and is required by
	// the planner to know whether or not he's actually communicating with a
	// status endpoint).
	AckConnect    bool `protobuf:"varint,2048,opt,name=ackConnect" json:"ackConnect,omitempty"`
	AckDisconnect bool `protobuf:"varint,2049,opt,name=ackDisconnect" json:"ackDisconnect,omitempty"`
	// These byte arrays are for any potential data transferred back
	// that cannot be foreseen here (e.g. data for debugging purposes)
	ExtraData [][]byte `protobuf:"bytes,2050,rep,name=extraData,proto3" json:"extraData,omitempty"`
}

func (m *ControlMessage) Reset()                    { *m = ControlMessage{} }
func (m *ControlMessage) String() string            { return proto.CompactTextString(m) }
func (*ControlMessage) ProtoMessage()               {}
func (*ControlMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func init() {
	proto.RegisterType((*StatusMessage)(nil), "statusMessagesProtobuf.StatusMessage")
	proto.RegisterType((*PlannerInfo)(nil), "statusMessagesProtobuf.PlannerInfo")
	proto.RegisterType((*SystemStatus)(nil), "statusMessagesProtobuf.SystemStatus")
	proto.RegisterType((*NetworkStatus)(nil), "statusMessagesProtobuf.NetworkStatus")
	proto.RegisterType((*PlannerStatus)(nil), "statusMessagesProtobuf.PlannerStatus")
	proto.RegisterType((*ServiceStatus)(nil), "statusMessagesProtobuf.ServiceStatus")
	proto.RegisterType((*Harddrive)(nil), "statusMessagesProtobuf.Harddrive")
	proto.RegisterType((*NetworkInterface)(nil), "statusMessagesProtobuf.NetworkInterface")
	proto.RegisterType((*StatusKvPair)(nil), "statusMessagesProtobuf.StatusKvPair")
	proto.RegisterType((*ControlMessage)(nil), "statusMessagesProtobuf.ControlMessage")
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 859 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x55, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0x06, 0x25, 0xf9, 0x47, 0x23, 0xc9, 0x70, 0x17, 0x85, 0xc1, 0x16, 0xfd, 0x71, 0xd9, 0x18,
	0xc8, 0xa1, 0x10, 0xd0, 0x16, 0xee, 0xb5, 0x88, 0x13, 0x14, 0x31, 0xd2, 0xa4, 0xc2, 0x3a, 0x41,
	0xd1, 0xe3, 0x5a, 0x5a, 0xc9, 0x84, 0x49, 0x2e, 0xb1, 0xbb, 0x74, 0xa2, 0x5b, 0xd3, 0xbe, 0x42,
	0x2f, 0x7d, 0x86, 0xa2, 0xaf, 0xd6, 0x67, 0xe8, 0xcc, 0x72, 0x49, 0x2d, 0xed, 0xc8, 0x6d, 0x6e,
	0x33, 0xdf, 0xfc, 0x70, 0x38, 0xf3, 0xcd, 0x2c, 0x1c, 0xe4, 0xd2, 0x18, 0xb1, 0x92, 0x66, 0x5a,
	0x6a, 0x65, 0x15, 0x3b, 0x32, 0x56, 0xd8, 0xca, 0x3c, 0xf7, 0xe8, 0x8c, 0xc0, 0xcb, 0x6a, 0x99,
	0xfc, 0xd3, 0x87, 0xc9, 0x45, 0x68, 0x62, 0x87, 0xd0, 0x2f, 0xd3, 0x45, 0xbc, 0x73, 0x1c, 0x3d,
	0x1c, 0x70, 0x12, 0x19, 0x83, 0x41, 0x55, 0x21, 0xb4, 0x8b, 0xd0, 0x98, 0x3b, 0x99, 0x1d, 0xc3,
	0x28, 0x17, 0xf3, 0xab, 0xb4, 0x90, 0xaf, 0xc8, 0xb4, 0xe7, 0x4c, 0x21, 0xc4, 0x3e, 0x81, 0xa1,
	0x4d, 0xb1, 0x0a, 0x2b, 0xf2, 0x32, 0xde, 0x77, 0xd9, 0x36, 0x00, 0xfb, 0x01, 0x46, 0x65, 0x26,
	0x8a, 0x42, 0xea, 0xf3, 0x62, 0xa9, 0xe2, 0x5f, 0x0f, 0xd1, 0x61, 0xf4, 0xcd, 0x97, 0xd3, 0x77,
	0x97, 0x39, 0x9d, 0x6d, 0x7c, 0x79, 0x18, 0xc8, 0x9e, 0xc2, 0xd8, 0xac, 0x8d, 0x95, 0x79, 0xfd,
	0x13, 0x71, 0xe4, 0xf2, 0x3c, 0xd8, 0x96, 0xe7, 0x22, 0xf0, 0xe5, 0x9d, 0x48, 0xf6, 0x0c, 0x26,
	0x85, 0xb4, 0xaf, 0x95, 0xbe, 0xf6, 0xa9, 0x7a, 0x2e, 0xd5, 0xc9, 0xb6, 0x54, 0x2f, 0x42, 0x67,
	0xde, 0x8d, 0xa5, 0x64, 0xbe, 0x4a, 0x9f, 0xac, 0x7f, 0x7f, 0xb2, 0x59, 0xe8, 0xcc, 0xbb, 0xb1,
	0x94, 0xcc, 0x48, 0x7d, 0x93, 0xce, 0xa5, 0x4f, 0x36, 0xb8, 0x3f, 0xd9, 0x45, 0xe8, 0xcc, 0xbb,
	0xb1, 0xc9, 0xef, 0x11, 0x8c, 0x82, 0x6e, 0xd2, 0x70, 0x0b, 0x91, 0x4b, 0xd7, 0xb8, 0x21, 0x77,
	0x32, 0x7b, 0x00, 0x93, 0x2c, 0xc5, 0xd6, 0x14, 0x8f, 0x16, 0x0b, 0x8d, 0xd9, 0x5d, 0xf5, 0x43,
	0xde, 0x05, 0xd9, 0xe7, 0x00, 0x8b, 0xd4, 0xcc, 0x15, 0xa6, 0x9a, 0xdb, 0x7a, 0x82, 0xfb, 0x3c,
	0x80, 0xd8, 0x47, 0xb0, 0xd7, 0x58, 0xdf, 0xd6, 0xd6, 0x46, 0x4f, 0xfe, 0xec, 0xc3, 0x38, 0x9c,
	0x05, 0x3b, 0x82, 0xdd, 0xaa, 0x24, 0x7a, 0xb8, 0x42, 0x06, 0xdc, 0x6b, 0xc4, 0xa2, 0x79, 0x59,
	0x9d, 0xff, 0xf4, 0xb3, 0x48, 0xad, 0x9b, 0x08, 0xb2, 0xa8, 0x05, 0x58, 0x8c, 0x5f, 0x40, 0x65,
	0x91, 0x49, 0x57, 0xe2, 0x80, 0x37, 0xaa, 0xb7, 0x9c, 0x55, 0x66, 0xed, 0xba, 0x55, 0x5b, 0x48,
	0x65, 0x1f, 0xc3, 0x3e, 0x8a, 0x2f, 0x95, 0x15, 0x99, 0x27, 0x79, 0xab, 0x3b, 0x56, 0xcb, 0x5c,
	0xe9, 0xf5, 0x2b, 0xea, 0xa8, 0x23, 0xfc, 0x80, 0x87, 0x10, 0xd5, 0x53, 0xab, 0xcf, 0xc5, 0x1b,
	0xc7, 0x7a, 0xac, 0xa7, 0x05, 0xc8, 0x6a, 0x5e, 0x8b, 0xb2, 0x8e, 0xf6, 0x9c, 0x6f, 0x01, 0xaa,
	0x89, 0x14, 0x8a, 0x1c, 0xd6, 0x35, 0x79, 0x95, 0x3d, 0x02, 0xb8, 0x12, 0x1a, 0xfb, 0x9a, 0xde,
	0x48, 0x13, 0xc3, 0x71, 0x1f, 0xc7, 0xfb, 0xc5, 0xb6, 0xf1, 0x3e, 0x6d, 0x3c, 0x79, 0x10, 0x44,
	0x0d, 0xfc, 0x51, 0x89, 0x85, 0xf9, 0x3a, 0x1e, 0x61, 0xee, 0x88, 0xef, 0x66, 0x4e, 0x6b, 0xf1,
	0xd3, 0x78, 0x1c, 0xe0, 0xa7, 0x54, 0x4c, 0xed, 0x7f, 0x1a, 0x4f, 0x9c, 0x61, 0xaf, 0x0e, 0x38,
	0x4d, 0x7e, 0x81, 0x49, 0x87, 0xdb, 0xb8, 0x63, 0x90, 0x16, 0x56, 0xea, 0xa5, 0x98, 0x4b, 0xda,
	0x30, 0xaa, 0xee, 0xe1, 0x7f, 0xac, 0xc5, 0x79, 0x13, 0xc0, 0x83, 0xd8, 0x44, 0xc1, 0xa4, 0xc3,
	0x74, 0xf6, 0x15, 0x7c, 0x80, 0x94, 0x58, 0xa6, 0x2b, 0x8c, 0x5f, 0xa6, 0x99, 0x7c, 0xb1, 0xa1,
	0xe2, 0x5d, 0x03, 0x71, 0x35, 0x53, 0x2b, 0xda, 0xcc, 0x3e, 0x71, 0x95, 0x64, 0x6a, 0xb9, 0x7c,
	0x63, 0xb5, 0x78, 0x22, 0xac, 0x88, 0x0f, 0xd1, 0x30, 0xe6, 0x1b, 0x20, 0xf9, 0x2b, 0xc2, 0xf3,
	0x16, 0xf2, 0xff, 0xfd, 0xbf, 0xe8, 0xb6, 0xa3, 0x17, 0x6c, 0x07, 0x62, 0xa5, 0xd2, 0xd6, 0x31,
	0x6e, 0xc2, 0x9d, 0x4c, 0x98, 0x15, 0xe6, 0xda, 0x71, 0x0d, 0xfd, 0x48, 0x6e, 0xab, 0xdd, 0xf9,
	0xdf, 0xd5, 0xbe, 0x8d, 0x60, 0xd8, 0x4e, 0x97, 0x26, 0xb7, 0x34, 0x2f, 0xd7, 0x65, 0x53, 0x9e,
	0xd7, 0xd8, 0x67, 0x00, 0xb9, 0xaa, 0x0a, 0x3b, 0x53, 0xd8, 0x59, 0x5f, 0x59, 0x80, 0xb8, 0x73,
	0x6d, 0xe4, 0xc2, 0x6f, 0x84, 0x93, 0xd9, 0x87, 0xb0, 0x63, 0x1d, 0xe3, 0xeb, 0x65, 0xa8, 0x15,
	0xf2, 0x5c, 0x6a, 0x29, 0xfd, 0x1a, 0x38, 0x39, 0xf9, 0x23, 0x82, 0xc3, 0xdb, 0x33, 0x64, 0x07,
	0xd0, 0xc3, 0x23, 0x4f, 0x65, 0xec, 0xf0, 0x5e, 0xfd, 0x22, 0xdc, 0x69, 0x0b, 0xf9, 0x94, 0xee,
	0xa3, 0x63, 0xf4, 0x29, 0x89, 0x60, 0x78, 0x13, 0xf3, 0xa6, 0x2b, 0x63, 0xde, 0xa8, 0xd4, 0x84,
	0x4b, 0x8d, 0x64, 0x9b, 0x0b, 0x63, 0xdd, 0xb7, 0xb1, 0x09, 0x2d, 0x40, 0xa5, 0xe2, 0x05, 0x29,
	0xeb, 0xed, 0x1b, 0xf2, 0x5a, 0x49, 0xbe, 0xc3, 0x7b, 0xe1, 0x06, 0xf8, 0xec, 0x66, 0x26, 0x52,
	0x4d, 0xaf, 0xd4, 0xb5, 0x5c, 0xfb, 0xce, 0x90, 0x48, 0x71, 0x37, 0x22, 0xab, 0x9a, 0xa2, 0x6a,
	0x25, 0xf9, 0x3b, 0x82, 0x83, 0xc7, 0xaa, 0xb0, 0x5a, 0x65, 0xcd, 0x03, 0xd7, 0x3c, 0x67, 0xd1,
	0xf6, 0xe7, 0xac, 0x77, 0xf7, 0x39, 0xc3, 0x6b, 0x27, 0xe6, 0xd7, 0x8f, 0x6f, 0x5d, 0xbb, 0x0d,
	0xc4, 0x4e, 0x60, 0x82, 0xda, 0x93, 0xcd, 0x45, 0xf4, 0x37, 0xaf, 0x8b, 0xb2, 0x4f, 0x43, 0x06,
	0xfc, 0x76, 0x9b, 0x02, 0x67, 0xdf, 0xc3, 0x89, 0xd2, 0xab, 0xe9, 0x95, 0xca, 0xf0, 0xa5, 0xc4,
	0xa7, 0x1b, 0xb7, 0xc6, 0xa4, 0xc5, 0x6a, 0x8a, 0x03, 0x93, 0xf9, 0x34, 0x57, 0x45, 0x6a, 0x95,
	0x46, 0xe0, 0xec, 0xe8, 0xe2, 0x9d, 0xfb, 0x77, 0xb9, 0xeb, 0xde, 0xfb, 0x6f, 0xff, 0x0d, 0x00,
	0x00, 0xff, 0xff, 0x3e, 0xce, 0xb8, 0x7c, 0x01, 0x08, 0x00, 0x00,
}
