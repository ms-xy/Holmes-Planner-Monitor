package main

import (
	statusclient "github.com/ms-xy/Holmes-Planner-Monitor/go/client"
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

// STRESS TEST CLIENT

var (
	waveLen     = 10 * time.Second
	waves       = 1
	waveSize    = 1
	services    = 1
	storageAddr = "10.0.4.79:9016"
	// storageAddr = "127.0.0.1:9016"
)

func main() {
	for i := 0; i < waves; i++ {
		fmt.Println("Launching Wave:", i)
		for j := 0; j < waveSize; j++ {
			go planner((i+1)*j, services, time.Duration(waves-i)*30*waveLen)
		}
		<-time.After(waveLen)
	}
	<-time.After(10 * time.Second) // disconnect grace period
}

var (
	// size: 1732 chars
	long_log_message = `This is a very long log message with multiple lines.
	aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
	bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
	cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
	dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
	eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
	ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
	gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg
	aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
	bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
	cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
	dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
	eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
	ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
	gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg
	aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
	bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
	cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
	dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
	eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
	ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
	gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg`
	// size: 10'392
	log_message_bundle = []string{long_log_message, long_log_message,
		long_log_message, long_log_message, long_log_message, long_log_message}
)

func planner(p, services int, d time.Duration) {
	// Adjust log level to debug for monitor
	monitor := statusclient.NewInstance()
	monitor.SetLogLevel(statusclient.LogLevelDebug)

	// Connect to a status server
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("10.0.4.33:%d", 7788+p))
	err := monitor.Connect(storageAddr, &msgtypes.PlannerInfo{
		Name:          "TestPlanner",
		ListenAddress: addr,
		Connect:       true,
	})
	check(err)

	// Send some planner status
	monitor.PlannerStatus("default-config", log_message_bundle, nil)

	for s := 0; s < services; s++ {
		go service(p, s, monitor)
	}

	// this is just additional and not needed (here for evaluation purposes)
	// si, err := sysinfo.New()
	// fmt.Printf("memory usage: %s/%s \n", humanize.Bytes((si.Ram.Used)), humanize.Bytes((si.Ram.Total)))
	// fmt.Printf("swap usage: %s/%s \n", humanize.Bytes((si.Swap.Used)), humanize.Bytes((si.Swap.Total)))

	// for msg := range monitor.IncomingControlMessages() {
	// 	fmt.Println("received control message: ", msg)
	// }

	time.Sleep(d)
	monitor.Disconnect()
}

func service(p, s int, m *statusclient.Monitor) {
	for {
		m.Logf(statusclient.LogLevelInfo, "sending message for %d__%d", p, s)
		m.ServiceStatus(
			"default-config",
			log_message_bundle,
			nil,
			fmt.Sprintf("service-%d", s),
			uint16(7700+s*10),
			"hello_world",
		)

		// send every second
		<-time.After(1 * time.Second)
	}
}
