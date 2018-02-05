package main

import (
	"fmt"
	"github.com/tarm/serial"
	"sync"
	"time"
)

var AcquireArduinoUID = func() func(byte) *serial.Port {
	mtx := sync.Mutex{}
	return func(uid byte) *serial.Port {
		mtx.Lock()
		defer mtx.Unlock()
		for i := range []int{0, 1, 2} {
			serialConf := &serial.Config{Name: fmt.Sprint("/dev/ttyACM", i), Baud: 9600, ReadTimeout: time.Second * 5}
			serialPort, err := serial.OpenPort(serialConf) // Try to open a serial port.
			if err == nil {
				serialPort.Write([]byte{identReq})
				serialPort.Flush()
				buf := make([]byte, 1)
				count, err := serialPort.Read(buf)
				serialPort.Close()
				if err != nil {
					fmt.Println("Error in reading from", uid, ":", err)
				} else if count != 1 {
					fmt.Println("Reading from Arduino timed out")
				} else if buf[0] == uid {
					return serialPort
				} else {
					fmt.Println("Wrong port for", uid)
				}
			} else {
				fmt.Println("Failed to acquire serial port", uid, ":", err)
			}
		}
		return nil
	}
}()
