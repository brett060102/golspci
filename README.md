# golspci

`golspci` is a Go library that parses the output of the `lspci` command, make it easier to collect hardware info by program.  
This library will call `lspci -vmm -D` to get canonical output for easier parsing.

## Note

`lspci` provides `-n` flag, which show PCI vendor and device codes as numbers instead of looking them up in the [PCI ID list](https://pci-ids.ucw.cz/v2.2/pci.ids). The code never change between different version of `lspci`, but name does. Pass vendorInNumber=true when calling `lspci.New` to get codes instead of name.

## Usage

```go
package main

import (
    "fmt"
    "github.com/tmojzes/golspci"
)

func main() {
    // vendorInNumber=false to get text version of vendor
    // vendorInNumber=true to get codes version of vendor
    l := golspci.New(false)

    if err := l.Parse(); err != nil {
        panic(err)
    }

	for _, device := range l.Devices {
		fmt.Println(device.Name)
	}
    // You will get something like:
    /*
        Tiger Lake-LP Serial IO I2C Controller #0
        Tiger Lake-LP LPC Controller
        Tiger Lake-LP Smart Sound Technology Audio Controller
        Tiger Lake-LP Thunderbolt 4 NHI #0
        ...
     */
}
```
