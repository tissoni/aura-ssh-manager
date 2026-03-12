package utils

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strings"
)

type Process struct {
	Name string
	PID  string
	Port string
}

func ListActivePorts() ([]Process, error) {
	// lsof -i -P -n | grep LISTEN
	cmd := exec.Command("lsof", "-i", "-P", "-n")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var processes []Process
	scanner := bufio.NewScanner(&out)
	
	// Skip header
	if scanner.Scan() {
		// Header: COMMAND  PID  USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
	}

	portRegex := regexp.MustCompile(`:(\d+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "(LISTEN)") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		name := fields[0]
		pid := fields[1]
		address := fields[8]

		matches := portRegex.FindStringSubmatch(address)
		port := "unknown"
		if len(matches) > 1 {
			port = matches[1]
		}

		processes = append(processes, Process{
			Name: name,
			PID:  pid,
			Port: port,
		})
	}

	return processes, nil
}

func KillProcess(pid string) error {
	return exec.Command("kill", "-9", pid).Run()
}
