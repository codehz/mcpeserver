package main

import (
	"errors"
	"time"
)

func runExec(profile, command string, timeout int) (string, error) {
	var bus bus
	bus.init(profile)
	defer bus.close()
	id := ^uint32(0)
	errchan := make(chan error)
	timer := time.NewTimer(time.Duration(timeout) * time.Millisecond)
	cached := make(map[uint32]string)
	resultchan := make(chan string)

	go func() {
		rid, err := bus.exec(command)
		if err != nil {
			errchan <- err
		} else {
			id = rid
			if res, ok := cached[id]; ok {
				resultchan <- res
			}
		}
	}()

	for {
		select {
		case <-timer.C:
			return "", errors.New("Timout")
		case v := <-bus.log:
			if v.Name == "one.codehz.bedrockserver.core.exec_result" {
				if v.Body[0].(uint32) == id {
					return v.Body[1].(string), nil
				}
			}
		case res := <-resultchan:
			return res, nil
		}
	}
	return "", nil
}
