package main

func runExec(profile, command string, timeout int) (string, error) {
	var bus bus
	bus.init(profile)
	defer bus.close()

	return bus.exec(command)
}
