package driver  // where "driver" is the folder that contains io.go, io.c, io.h, channels.go, channels.c and driver.go
/*
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/
import "C"

//func Io_init(type ElevatorType) bool {
	//return bool(int(C.io_init(ElevatorType type)) != 1)
//}
func Io_init() bool {
	return bool(int(C.io_init(/*C.ElevatorType(type)*/)) != 1)
}

func Io_set_bit(channel int) {
	C.io_set_bit(C.int(channel))
}

func Io_clear_bit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func Io_write_analog(channel, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}

func Io_read_bit(channel int) bool {
	return int(C.io_read_bit(C.int(channel))) != 0
}

func Io_read_analog(channel int) int {
	return int(C.io_read_analog(C.int(channel)))
}
