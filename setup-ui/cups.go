package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type printer struct {
	Name        string
	DisplayName string
	State       string
	StateClass  string
	URI         string
	Jobs        int
}

type device struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

type model struct {
	URI  string
	Name string
}

func listPrinters() ([]printer, error) {
	statesOutput, err := runCommand("lpstat", "-p")
	if err != nil {
		if strings.Contains(err.Error(), "No destinations added") {
			return []printer{}, nil
		}
		return nil, err
	}
	uriOutput, _ := runCommand("lpstat", "-v")
	descriptionOutput, _ := runCommand("lpstat", "-l", "-p")
	jobsOutput, _ := runCommand("lpstat", "-W", "not-completed", "-o")

	uris := parsePrinterURIs(uriOutput)
	descriptions := parseDescriptions(descriptionOutput)
	jobCounts := parseJobCounts(jobsOutput)
	var printers []printer

	scanner := bufio.NewScanner(strings.NewReader(statesOutput))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "printer ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name := fields[1]
		state, stateClass := friendlyPrinterState(line)
		displayName := descriptions[name]
		if displayName == "" {
			displayName = strings.ReplaceAll(name, "_", " ")
		}
		printers = append(printers, printer{
			Name:        name,
			DisplayName: displayName,
			State:       state,
			StateClass:  stateClass,
			URI:         uris[name],
			Jobs:        jobCounts[name],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.Slice(printers, func(i, j int) bool { return printers[i].DisplayName < printers[j].DisplayName })
	return printers, nil
}

func discoverDevices() ([]device, error) {
	output, err := runCommandWithTimeout(12*time.Second, "lpinfo", "-v")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	devices := make([]device, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(fields) < 2 {
			continue
		}
		uri := fields[1]
		// lpinfo lists both usable device URIs and bare backend names such as
		// "ipp" or "socket". Only complete URIs represent an actual printer.
		if !strings.Contains(uri, "://") {
			continue
		}
		protocol := strings.ToLower(strings.SplitN(uri, ":", 2)[0])
		if protocol != "dnssd" && protocol != "socket" && protocol != "lpd" && protocol != "ipp" && protocol != "ipps" {
			continue
		}
		name := friendlyDeviceName(uri)
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		devices = append(devices, device{URI: uri, Name: name, Protocol: strings.ToUpper(protocol)})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(devices, func(i, j int) bool {
		if strings.HasPrefix(devices[i].URI, "dnssd:") != strings.HasPrefix(devices[j].URI, "dnssd:") {
			return strings.HasPrefix(devices[i].URI, "dnssd:")
		}
		return devices[i].Name < devices[j].Name
	})
	return preferFriendlyDevices(devices), nil
}

func preferFriendlyDevices(devices []device) []device {
	hasDNSSD := false
	for _, item := range devices {
		if strings.EqualFold(item.Protocol, "DNSSD") {
			hasDNSSD = true
			break
		}
	}
	if !hasDNSSD {
		return devices
	}

	// Many Brother devices announce the same physical printer both with a
	// readable DNS-SD name and a cryptic BRN… LPD queue. If DNS-SD works, hide
	// those duplicates from the simple UI; manual IP setup remains available.
	filtered := make([]device, 0, len(devices))
	for _, item := range devices {
		if !strings.EqualFold(item.Protocol, "LPD") {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func listModels() ([]model, error) {
	output, err := runCommand("lpinfo", "-m")
	if err != nil {
		return nil, err
	}
	var models []model
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(strings.ToLower(line), "brlaser") {
			continue
		}
		uri, name, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		name = cleanModelName(name)
		models = append(models, model{URI: uri, Name: name})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sortModels(models)
	return models, nil
}

func runCommand(name string, args ...string) (string, error) {
	return runCommandWithTimeout(15*time.Second, name, args...)
}

func runCommandWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	timer := time.AfterFunc(timeout, func() {
		if command.Process != nil {
			_ = command.Process.Kill()
		}
	})
	defer timer.Stop()
	output, err := command.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func parsePrinterURIs(output string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.TrimPrefix(line, "device for ")
		name, uri, found := strings.Cut(line, ": ")
		if found {
			result[name] = uri
		}
	}
	return result
}

func parseDescriptions(output string) map[string]string {
	result := make(map[string]string)
	current := ""
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "printer ") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				current = fields[1]
			}
			continue
		}
		if current != "" && strings.HasPrefix(line, "Description: ") {
			result[current] = strings.TrimPrefix(line, "Description: ")
		}
	}
	return result
}

func parseJobCounts(output string) map[string]int {
	result := make(map[string]int)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		jobID := strings.Fields(line)[0]
		if split := strings.LastIndex(jobID, "-"); split > 0 {
			result[jobID[:split]]++
		}
	}
	return result
}

// friendlyPrinterState returns a message key (see i18n.go) and a CSS class.
func friendlyPrinterState(line string) (string, string) {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "disabled"):
		return "state.stopped", "warning"
	case strings.Contains(lower, "printing"):
		return "state.printing", "working"
	case strings.Contains(lower, "idle"):
		return "state.ready", "ready"
	default:
		return "state.unknown", "neutral"
	}
}

func friendlyDeviceName(uri string) string {
	withoutScheme := uri
	if _, rest, found := strings.Cut(uri, "://"); found {
		withoutScheme = rest
	}
	if decoded, err := urlQueryUnescape(withoutScheme); err == nil {
		withoutScheme = decoded
	}
	if index := strings.Index(withoutScheme, "._"); index > 0 {
		withoutScheme = withoutScheme[:index]
	}
	withoutScheme = strings.Trim(withoutScheme, "/")
	withoutScheme = strings.ReplaceAll(withoutScheme, "_", " ")
	if withoutScheme == "" {
		return uri
	}
	return withoutScheme
}

func urlQueryUnescape(value string) (string, error) {
	value = strings.ReplaceAll(value, "+", "%2B")
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value, err
	}
	return decoded, nil
}

func cleanModelName(name string) string {
	name = strings.TrimSpace(name)
	if index := strings.Index(strings.ToLower(name), ", using "); index > 0 {
		name = name[:index]
	}
	if index := strings.Index(strings.ToLower(name), " using "); index > 0 {
		name = name[:index]
	}
	return name
}
