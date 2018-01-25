package main

import (
	"fmt"
	"github.com/tarm/serial"
	"sync"
)

var AcquireArduinoUID = func() func(byte) *serial.Port {
	mtx := sync.Mutex{}
	return func(uid byte) *serial.Port {
		mtx.Lock()
		defer mtx.Unlock()
		for i := range []int{0, 1, 2} {
			serialConf := &serial.Config{Name: fmt.Sprint("/dev/ttyACM", i), Baud: 9600}
			serialPort, err := serial.OpenPort(serialConf) // Try to open a serial port.
			if err == nil {
				serialPort.Write([]byte{identReq})
				buf := make([]byte, 1)
				serialPort.Read(buf)
				if buf[0] == uid {
					return serialPort
				}
				serialPort.Close()
				fmt.Println("Wrong port for", uid)
			} else {
				fmt.Println("Failed to acquire serial port:", uid, err)
			}
		}
		return nil
	}
}()
