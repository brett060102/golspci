package golspci

import (
	"bufio"
	"bytes"
	"fmt"
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
	slot     string
	class    string
	vendor   string
	device   string
	sVendor  string
	sDevice  string
	rev      int
	progIf   int
	numaNode int
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
	for _, device := range devices {
		pciDevice := pciDevice{}
		for k, v := range device {
			switch k {
			case "Slot":
				pciDevice.slot = v
			case "Class":
				pciDevice.class = v
			case "Vendor":
				pciDevice.vendor = v
			case "Device":
				pciDevice.device = v
			case "SVendor":
				pciDevice.sVendor = v
			case "SDevice":
				pciDevice.sDevice = v
			case "Rev":
				rev, err := strconv.Atoi(v)
				if err != nil {
					log.Fatalf("Failed to convert value: %v to int. Got error: %v", v, err)
				}
				pciDevice.rev = rev
			case "ProgIf":
				progIf, err := strconv.Atoi(v)
				if err != nil {
					log.Fatalf("Failed to convert value: %v to int. Got error: %v", v, err)
				}
				pciDevice.progIf = progIf
			case "NUMANode":
				fmt.Println(k)
				fmt.Println(v)
				numaNode, err := strconv.Atoi(v)
				if err != nil {
					log.Fatalf("Failed to convert value: %v to int. Got error: %v", v, err)
				}
				pciDevice.numaNode = numaNode
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
