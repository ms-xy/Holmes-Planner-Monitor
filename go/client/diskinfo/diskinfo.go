// +build linux

package diskinfo

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/ms-xy/Holmes-Planner-Monitor/go/msgtypes"
)

var (
	//                               1        2         3         4         5         6
	//                               fstype   total     used      free      use%      mount-point
	re_record = regexp.MustCompile("([^ ]+) +([0-9]+) +([0-9]+) +([0-9]+) +([^%]+)% +(.+)")
)

func Get() ([]*msgtypes.Harddrive, error) {
	var (
		// data acquiration
		data     string
		splitted []string
		n        int
		result   []*msgtypes.Harddrive
		part     string
		record   []string
		// processing
		total uint64
		used  uint64
		free  uint64
		// refining

		// errors
		err error
	)

	data, err = run_df_cmd()
	if err != nil {
		return nil, err
	}

	splitted = strings.Split(data, "\n")
	splitted = splitted[1:] // skip the first line (headers)

	n = 0
	result = make([]*msgtypes.Harddrive, len(splitted))

	for _, part = range splitted {

		if part != "" {

			record = re_record.FindStringSubmatch(part)
			// 1        2         3         4         5         6
			// fstype   total     used      free      use%      mount-point

			if total, err = strconv.ParseUint(record[2], 10, 64); err != nil {
				return nil, err
			}
			if used, err = strconv.ParseUint(record[3], 10, 64); err != nil {
				return nil, err
			}
			if free, err = strconv.ParseUint(record[4], 10, 64); err != nil {
				return nil, err
			}
			if _, err = strconv.ParseUint(record[5], 10, 64); err != nil {
				return nil, err
			}

			result[n] = &msgtypes.Harddrive{
				FsType:     record[1],
				MountPoint: record[6],
				Total:      total * 1024,
				Used:       used * 1024,
				Free:       free * 1024,
			}

			n++

		}
	}

	return result[:n], nil
}

func run_df_cmd() (string, error) {
	// This requires the `df` command to be installed.
	// The size is reported in 1024 byte blocks.
	cmd := exec.Command("df")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	cmd.Start()

	buf := make([]byte, 0x8000) // ~32KiB should be enough
	n, err := stdout.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}
