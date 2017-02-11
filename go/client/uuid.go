package client

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (this *Monitor) load_machine_uuid() error {
	// Create the folder, so we don't get a needless error
	err := os.MkdirAll(filepath.Dir(this.uuid_file_path), 0700)
	if err != nil {
		return err
	}

	// Use os.Stat() to determine whether or not the file exists (and is no dir)
	if fi, err := os.Stat(this.uuid_file_path); err == nil {
		if fi.IsDir() {
			return errors.New(this.uuid_file_path + ": expected a file, found a directory")
		}
		bytes, err := ioutil.ReadFile(this.uuid_file_path)
		if err != nil {
			return err
		}
		if len(bytes) > 0 {
			err := this.machine_uuid.FromString(string(bytes))
			if err != nil {
				return errors.New(this.uuid_file_path + " contains a malformed uuid: " + err.Error())
			}
			this.Logf(LogLevelDebug, "MachineUUID=%s", this.machine_uuid.ToString())
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (this *Monitor) save_machine_uuid() error {
	// Write uuid file
	// TODO: what to do in case of different previous content in this file?
	//       That would be a severe error ...
	return ioutil.WriteFile(this.uuid_file_path, []byte(this.machine_uuid.ToString()), 0600)
}
