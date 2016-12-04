package main

import (
	monitor "github.com/ms-xy/Holmes-Planner-Monitor/go/client"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"

	// "github.com/dustin/go-humanize"
	// "github.com/ms-xy/Holmes-Planner-Monitor/go/client/sysinfo"

	"fmt"
	"net"
	"time"
)

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	// Adjust log level to debug for monitor
	monitor.SetLogLevel(monitor.LogLevelDebug)
	// Connect to a local status server
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8016")
	err := monitor.Connect("127.0.0.1:9016", &msgtypes.PlannerInfo{
		Name:          "TestPlanner",
		ListenAddress: addr,
		Connect:       true,
	})
	check(err)
	// Send some planner status

	// // Old method:
	// monitor.PlannerStatus(&msgtypes.PlannerStatus{
	// 	ConfigProfileName: "default",
	// })

	// // New method
	monitor.PlannerStatus("default-config", []string{"Startup complete"}, nil)

	for a := 0; a < 5; a++ {
		go service(a)
	}

	// this is just additional and not needed (here for evaluation purposes)
	// si, err := sysinfo.New()
	// fmt.Printf("memory usage: %s/%s \n", humanize.Bytes((si.Ram.Used)), humanize.Bytes((si.Ram.Total)))
	// fmt.Printf("swap usage: %s/%s \n", humanize.Bytes((si.Swap.Used)), humanize.Bytes((si.Swap.Total)))

	for msg := range monitor.IncomingControlMessages() {
		fmt.Println("received control message: ", msg)
	}

	// monitor.Disconnect()
	// time.Sleep(10 * time.Second)
}

func service(a int) {
	for {
		// // Old method
		// monitor.ServiceStatus(&msgtypes.ServiceStatus{
		// 	ConfigProfileName: "Default_Config",
		// 	Name:              fmt.Sprintf("Service-%d", a),
		// 	Port:              uint16(7700 + a*10),
		//  Task:							 "hello_world",
		// })

		// // New method
		monitor.ServiceStatus("default-config", nil, nil, fmt.Sprintf("service-%d", a), uint16(7700+a*10), "hello_world")

		// send every 7 seconds
		<-time.After(7 * time.Second)
	}
}
