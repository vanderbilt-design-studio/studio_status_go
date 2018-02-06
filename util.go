package main

import (
	"fmt"
	"github.com/tarm/serial"
	"sync"
	"time"
)

var AcquireArduinoUID = func() func(byte) *serial.Port {
	mtx := sync.Mutex{}
	acquiredMap := make(map[int]bool)
	for _, port := range []int{0, 1, 2} {
		acquiredMap[port] = false
	}
	return func(uid byte) *serial.Port {
		mtx.Lock()
		defer mtx.Unlock()
		for port, isAcquired := range acquiredMap {
			if isAcquired {
				continue
			}
			serialConf := &serial.Config{Name: fmt.Sprint("/dev/ttyACM", port), Baud: 9600, ReadTimeout: time.Second * 5}
			serialPort, err := serial.OpenPort(serialConf) // Try to open a serial port.
			if err == nil {
				serialPort.Write([]byte{identReq})
				serialPort.Flush()
				buf := make([]byte, 1)
				count, err := serialPort.Read(buf)
				if buf[0] == uid && err != nil {
					fmt.Println("Acquired", uid)
					acquiredMap[port] = true
					return serialPort
				}
				serialPort.Close()
				if err != nil {
					fmt.Println("Error in reading from", uid, ":", err)
				} else if count != 1 {
					fmt.Println("Reading from Arduino timed out")
				} else {
					fmt.Println("Wrong port for", uid)
				}
			} else {
				fmt.Println("Failed to acquire serial port", uid, ":", err)
				acquiredMap[port] = true
			}
		}
		return nil
	}
}()
