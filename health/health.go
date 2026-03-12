package health

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
)

var httpClient = &http.Client{
	Timeout: 3 * time.Second,
}

func Check(hostname, port string) Status {
	if port == "" {
		port = "22"
	}
	address := net.JoinHostPort(hostname, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return StatusDown
	}
	conn.Close()
	return StatusUp
}

func CheckHTTP(url string) Status {
	resp, err := httpClient.Get(url)
	if err != nil {
		return StatusDown
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return StatusUp
	}
	return StatusDown
}

func CheckAll(hosts map[string]string) map[string]Status {
	results := make(map[string]Status)
	// For now keeping it simple, can be concurrent later if needed
	for key, addr := range hosts {
		host, port, _ := net.SplitHostPort(addr)
		if host == "" {
			host = addr
		}
		results[key] = Check(host, port)
	}
	return results
}

var pingRegexp = regexp.MustCompile(`time=([\d.]+) ms`)

func Ping(hostname string) (time.Duration, error) {
	// Use -c 1 for 1 packet, -W 1 for 1 second timeout
	out, err := exec.Command("ping", "-c", "1", "-W", "1", hostname).Output()
	if err != nil {
		return 0, err
	}

	match := pingRegexp.FindStringSubmatch(string(out))
	if len(match) > 1 {
		ms, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(ms * float64(time.Millisecond)), nil
	}

	return 0, nil
}

type CheckResult struct {
	Key    string
	Status Status
}

func CheckConcurrent(hosts []struct{Key, Hostname, Port string}) <-chan CheckResult {
	out := make(chan CheckResult, len(hosts))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	
	go func() {
		defer cancel()
		defer close(out)
		
		sem := make(chan struct{}, 10) // Limit concurrency
		for _, h := range hosts {
			go func(h struct{Key, Hostname, Port string}) {
				sem <- struct{}{}
				defer func() { <-sem }()
				
				status := Check(h.Hostname, h.Port)
				select {
				case out <- CheckResult{Key: h.Key, Status: status}:
				case <-ctx.Done():
				}
			}(h)
		}
		
		// Wait for all or timeout
		for i := 0; i < len(hosts); i++ {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
	
	return out
}
type Diagnostic struct {
	LoadAvg string
	MemUsed string
	MemTotal string
	DiskUsed string
	DiskTotal string
	Uptime   string
}

func ParseDiagnostic(uptimeOut, freeOut, dfOut string) Diagnostic {
	d := Diagnostic{}
	
	// Parse Uptime & Load (e.g. " 23:58:00 up  4:55,  0 users,  load average: 0.00, 0.01, 0.05")
	if strings.Contains(uptimeOut, "load average:") {
		parts := strings.Split(uptimeOut, "load average:")
		if len(parts) > 1 {
			d.LoadAvg = strings.TrimSpace(parts[1])
		}
		upParts := strings.Split(uptimeOut, "up")
		if len(upParts) > 1 {
			d.Uptime = strings.TrimSpace(strings.Split(upParts[1], ",")[0])
		}
	}

	// Parse Free (e.g. "Mem:           15821        6328        2523          79        6968        9076")
	lines := strings.Split(freeOut, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				d.MemTotal = fields[1] + "MB"
				d.MemUsed = fields[2] + "MB"
			}
		}
	}

	// Parse DF (e.g. "/dev/sda1        20G  4.5G   15G  24% /")
	dfLines := strings.Split(dfOut, "\n")
	if len(dfLines) > 1 {
		fields := strings.Fields(dfLines[1])
		if len(fields) >= 4 {
			d.DiskTotal = fields[1]
			d.DiskUsed = fields[2]
		}
	}

	return d
}
