package golspci

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type LSPCI struct {
	Devices    []pciDevice
	flagNumber bool
}

type pciDevice struct {
	Slot     string
	Class    string
	Vendor   string
	Name     string
	SVendor  string
	SDevice  string
	Rev      int
	ProgIf   int
	NumaNode int
}

func New(vendorInNumber bool) *LSPCI {
	return &LSPCI{
		Devices:    []pciDevice{},
		flagNumber: vendorInNumber,
	}
}

// func (l *LSPCI) SetDevice(i int)

func (l *LSPCI) Parse() error {
	devices, err := getDevices(l.flagNumber)
	if err != nil {
		log.Fatalf("Failed to get devices, because of the following error: %v", err)
	}
	i := 0
	pciDevice := pciDevice{}
	for _, device := range devices {
		for k, v := range device {
			switch k {
			case "Slot":
				pciDevice.Slot = v
			case "Class":
				pciDevice.Class = v
			case "Vendor":
				pciDevice.Vendor = v
			case "Device":
				pciDevice.Name = v
			case "SVendor":
				pciDevice.SVendor = v
			case "SDevice":
				pciDevice.SDevice = v
			case "Rev":
				rev, err := strconv.Atoi(v)
				if err != nil {
					log.Fatalf("Failed to convert value: %v to int. Got error: %v", v, err)
				}
				pciDevice.Rev = rev
			case "ProgIf":
				progIf, err := strconv.Atoi(v)
				if err != nil {
					log.Fatalf("Failed to convert value: %v to int. Got error: %v", v, err)
				}
				pciDevice.ProgIf = progIf
			case "NUMANode":
				numaNode, err := strconv.Atoi(v)
				if err != nil {
					log.Fatalf("Failed to convert value: %v to int. Got error: %v", v, err)
				}
				pciDevice.NumaNode = numaNode
			}
		}
		l.Devices = append(l.Devices, pciDevice)
		i++
	}
	return err
}

func getDevices(vendorInNumber bool) (map[string]map[string]string, error) {
	bin, findErr := findBin("lspci")
	if findErr != nil {
		return nil, findErr
	}
	args := []string{"-vmm", "-D"}
	if vendorInNumber {
		args = append(args, "-n")
	}
	cmd := exec.Command(bin, args...)

	out := &bytes.Buffer{}
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return parseLSPCI(out)
}

func scanDoubleNewLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	sep := []byte{'\n', '\n'}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, sep); i >= 0 {
		return i + 2, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func parseLSPCI(r io.Reader) (map[string]map[string]string, error) {
	ret := make(map[string]map[string]string)
	scanner := bufio.NewScanner(r)
	scanner.Split(scanDoubleNewLine)
	for scanner.Scan() {
		// Per sector
		section := make(map[string]string)
		subScanner := bufio.NewScanner(bytes.NewBuffer(scanner.Bytes()))
		for subScanner.Scan() {
			data := strings.SplitN(subScanner.Text(), ":\t", 2)
			section[data[0]] = data[1]
		}
		if err := subScanner.Err(); err != nil {
			return nil, err
		}
		ret[section["Slot"]] = section
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func findBin(binary string) (string, error) {
	return exec.LookPath(binary)
}
