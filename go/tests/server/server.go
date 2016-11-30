package main

// import (
// 	"github.com/golang/protobuf/proto"
// 	// "github.com/ms-xy/Holmes-Planner-Monitor/go/monitormsgs"
// 	pb "github.com/ms-xy/Holmes-Planner-Monitor/protobuf/generated-go"

// 	"fmt"
// 	"net"
// 	// "time"
// )

// func check(err error) {
// 	if err != nil {
// 		panic(err.Error())
// 	}
// }

// func main() {
// 	var (
// 		laddr      *net.UDPAddr
// 		connection *net.UDPConn
// 		err        error
// 		buffer     []byte = make([]byte, 0xfde8) // buffer of size 65000
// 	)
// 	// Start UDP server
// 	if laddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:9016"); err == nil {
// 		connection, err = net.ListenUDP("udp", laddr)
// 		fmt.Printf("Bound to %s\n", laddr.String())
// 	}
// 	check(err)

// 	i := 0
// 	for true {
// 		n, addr, _ := connection.ReadFromUDP(buffer)
// 		data := buffer[0:n]
// 		statusmsg := &pb.StatusMessage{}
// 		proto.Unmarshal(data, statusmsg)
// 		if statusmsg.PlannerInfo != nil {
// 			controlmsg := &pb.ControlMessage{AckPlannerInfo: true}
// 			bytes, err := proto.Marshal(controlmsg)
// 			check(err)
// 			fmt.Println("PlannerInfo Acknowledged")
// 			connection.WriteToUDP(bytes, addr)

// 		} else {
// 			controlmsg := &pb.ControlMessage{TestData: fmt.Sprintf("Echo %v", statusmsg)}
// 			bytes, err := proto.Marshal(controlmsg)
// 			check(err)
// 			connection.WriteToUDP(bytes, addr)
// 		}
// 		fmt.Printf("Received %d bytes from address %s: %v\n", n, addr.String(), statusmsg)
// 		i++
// 	}
// 	check(err)
// }
