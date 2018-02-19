package main

import (
	"fmt"
	"github.com/tarm/serial"
	"github.com/veandco/go-sdl2/sdl"
	"sync"
)

var AcquireArduinoUID = func() func(byte) *serial.Port {
	mtx := sync.Mutex{}
	ports := []int{0, 1, 2}
	acquiredMap := make(map[int]bool)
	for port := range ports {
		acquiredMap[port] = false
	}
	return func(uid byte) *serial.Port {
		mtx.Lock()
		defer mtx.Unlock()
		for port := range ports {
			if acquiredMap[port] {
				continue
			}
			serialConf := &serial.Config{Name: fmt.Sprint("/dev/ttyACM", port), Baud: 9600}
			serialPort, err := serial.OpenPort(serialConf) // Try to open a serial port.
			if err == nil {
				serialPort.Write([]byte{identReq})
				serialPort.Flush()
				buf := make([]byte, 1)
				count, err := serialPort.Read(buf)
				if buf[0] == uid && err == nil {
					acquiredMap[port] = true
					fmt.Println("Successfully got Arduino for req", uid)
					return serialPort
				}
				serialPort.Close()
				if err != nil {
					fmt.Println("Error in reading from", uid, ":", err)
				} else if count != 1 {
					fmt.Println("Reading from Arduino timed out")
				} else {
					fmt.Println("Wrong port for", uid, "got", buf[0])
				}
			} else {
				fmt.Println("Failed to acquire serial port", uid, ":", err)
				acquiredMap[port] = true
			}
		}
		fmt.Println("No port found for", uid)
		return nil
	}
}()

func colorToUint32(c sdl.Color) uint32 {
	return uint32(c.B) ^ (uint32(c.G) << 8) ^ (uint32(c.R) << 16) ^ (uint32(c.A) << 24)
}
