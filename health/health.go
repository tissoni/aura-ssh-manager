package health

import (
	"context"
	"net"
	"time"
)

type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
)

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
