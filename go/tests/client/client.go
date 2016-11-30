package main

import (
	monitor "github.com/ms-xy/Holmes-Planner-Monitor/go/client"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/client/sysinfo"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"

	"github.com/dustin/go-humanize"

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
	// Create a component monitor
	monitor.PlannerStatus(&msgtypes.PlannerStatus{
		ConfigProfileName: "default",
	})

	for a := 0; a < 5; a++ {
		go func(a int) {
			monitor.ServiceStatus(&msgtypes.ServiceStatus{
				ConfigProfileName: "Default_Config",
				Name:              fmt.Sprintf("Service-%d", a),
				Port:              uint16(7700 + a*10),
			})
		}(a)
	}
	time.Sleep(2 * time.Second)
	for a := 0; a < 5; a++ {
		go func(a int) {
			monitor.ServiceStatus(&msgtypes.ServiceStatus{
				ConfigProfileName: "Default_Config",
				Name:              fmt.Sprintf("Service-%d", a),
				Port:              uint16(7700 + a*10),
				Task:              "hello_world",
			})
		}(a)
	}

	si, err := sysinfo.New()
	fmt.Printf("memory usage: %s/%s \n", humanize.Bytes((si.Ram.Used)), humanize.Bytes((si.Ram.Total)))
	fmt.Printf("swap usage: %s/%s \n", humanize.Bytes((si.Swap.Used)), humanize.Bytes((si.Swap.Total)))

	for msg := range monitor.IncomingControlMessages() {
		fmt.Println("received control message: ", msg)
	}

	// monitor.Disconnect()
	// time.Sleep(10 * time.Second)
}
