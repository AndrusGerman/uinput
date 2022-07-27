package uinput

import (
	"os"
	"unsafe"
)

type absInfo struct {
	Maximum    int32
	Minimum    int32
	Fuzz       int32
	Flat       int32
	Resolution int32
}

const (
	// These constants are defined in include/uapi/linux/uinput.h.
	UinputIoctlBase = 'U' // UINPUT_IOCTL_BASE
	AbsSetupIoctl   = 4   // UI_ABS_SETUP
)

type iocDir uint

const (
	iocWrite iocDir = 1
	iocRead  iocDir = 2
)

const (
	iocNrBits   = 8
	iocTypeBits = 8

	iocSizeBits = 14
	iocDirBits  = 2

	iocNrShift   = 0
	iocTypeShift = iocNrShift + iocNrBits
	iocSizeShift = iocTypeShift + iocTypeBits
	iocDirShift  = iocSizeShift + iocSizeBits
)

// ioc returns an encoded ioctl request in the given direction for the supplied type
// (e.g. UINPUT_IOCTL_BASE), number (e.g. UI_DEV_CREATE), and data size.
// This is analogous to the _IOC C macro.
func ioc(dir iocDir, typ, nr uint, size uintptr) uint {
	return (uint(dir) << iocDirShift) | (typ << iocTypeShift) | (nr << iocNrShift) |
		(uint(size) << iocSizeShift)
}

// Iow returns an encoded write ioctl request. See ioc for arguments.
// This is analogous to the _IOW C macro.
func primitiveIow(typ, nr uint, size uintptr) uintptr {
	return uintptr(ioc(iocWrite, typ, nr, size))
}

// setAbsSetup makes multiple UI_ABS_SETUP ioctls to a uinput FD to configure a virtual device.
func setAbsSetup(f *os.File, axes map[uint16]absInfo) error {
	for code, info := range axes {
		uinputAbsSetup := struct {
			code       uint16
			value      int32
			minimum    int32
			maximum    int32
			fuzz       int32
			flat       int32
			resolution int32
		}{
			code:       uint16(code),
			value:      0,
			minimum:    info.Minimum,
			maximum:    info.Maximum,
			fuzz:       info.Fuzz,
			flat:       info.Flat,
			resolution: info.Resolution,
		}
		if err := ioctl(f, primitiveIow(UinputIoctlBase, AbsSetupIoctl, unsafe.Sizeof(uinputAbsSetup)), uintptr(unsafe.Pointer(&uinputAbsSetup))); err != nil {
			return err
		}
	}
	return nil
}
