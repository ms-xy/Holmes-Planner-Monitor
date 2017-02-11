// +build linux

package netinfo

import (
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
)

var (
	//                               1         2        3
	//                               id        name     options
	re_record = regexp.MustCompile("([0-9]+): ([^ ]+) +([^\\\\]+)")
)

func Get() ([]*msgtypes.NetworkInterface, error) {
	var (
		// data acquiration
		data     string
		splitted []string
		n        int
		result   []*msgtypes.NetworkInterface
		part     string
		record   []string
		// processing
		id      int64
		options []string
		i       int
		// refining
		ip        net.IP
		cidr      *net.IPNet
		broadcast net.IP
		scope     string
		// errors
		err error
	)

	data, err = run_ip_cmd()
	if err != nil {
		return nil, err
	}

	splitted = strings.Split(data, "\n")

	n = 0
	result = make([]*msgtypes.NetworkInterface, len(splitted))

	for _, part = range splitted {

		if part != "" {

			record = re_record.FindStringSubmatch(part)

			id, err = strconv.ParseInt(record[1], 10, 32)
			if err != nil {
				return nil, err
			}

			options = strings.Split(record[3], " ")
			for i = 0; i < len(options)-1; i++ {
				switch options[i] {
				case "inet":
					ip, cidr, err = net.ParseCIDR(options[i+1])
					if err != nil {
						return nil, err
					}
				case "brd":
					broadcast = net.ParseIP(options[i+1])
				case "scope":
					scope = options[i+1]
				}
			}

			result[n] = &msgtypes.NetworkInterface{
				ID:        int(id),
				Name:      record[2],
				IP:        ip,
				Netmask:   msgtypes.IPMask(cidr.Mask),
				Broadcast: broadcast,
				Scope:     scope,
			}

			n++

		}
	}

	return result[:n], nil
}

func run_ip_cmd() (string, error) {
	// This requires the `ip` command to be installed ...
	// is there an alternative and potentially less OS dependent way?
	cmd := exec.Command("ip", "-o", "addr")

	// stdin, err := cmd.StdinPipe()
	// if err != nil {
	// 	return "", err
	// }
	// stdin.Close()

	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	return "", err
	// }

	// cmd.Start()
	// cmd.Wait()

	// buf := make([]byte, 0x8000) // ~32KiB should be enough
	// n, err := stdout.Read(buf)
	// if err != nil {
	// 	return "", err
	// }

	output, err := cmd.Output()

	return string(output), err
}
