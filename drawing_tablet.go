package uinput

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

// A DrawingTablet is a hybrid key / absolute change event output device.
// Is is used to enable a priogram to simulate DrawingTablet input events.
type DrawingTablet interface {
	// KeyPress will cause the key to be pressed and immediately released.
	ButtonPress(key int) error

	// ButtonDown will send a keypress event to an existing DrawingTablet device.
	// The key can be any of the predefined keycodes from keycodes.go.
	// Note that the key will be "held down" until "KeyUp" is called.
	ButtonDown(key int) error

	// ButtonUp will send a keyrelease event to an existing DrawingTablet device.
	// The key can be any of the predefined keycodes from keycodes.go.
	ButtonUp(key int) error

	// Sends out a SYN event.
	syncEvents() error

	Move(x, y float32) error

	Pressure(pressure float32) error

	io.Closer
}

type vDrawingTablet struct {
	name       []byte
	deviceFile *os.File
}

// CreateDrawingTablet will create a new DrawingTablet using the given uinput
// device path of the uinput device.
func CreateDrawingTablet(path string, name []byte) (DrawingTablet, error) { // TODO: Consider moving this to a generic function that works for all devices
	err := validateDevicePath(path)
	if err != nil {
		return nil, err
	}
	err = validateUinputName(name)
	if err != nil {
		return nil, err
	}

	fd, err := createVDrawingTabletDevice(path, name)
	if err != nil {
		return nil, err
	}

	return vDrawingTablet{name: name, deviceFile: fd}, nil
}

func (vg vDrawingTablet) ButtonPress(key int) error {
	err := vg.ButtonDown(key)
	if err != nil {
		return err
	}
	err = vg.ButtonUp(key)
	if err != nil {
		return err
	}
	return nil
}

func (vg vDrawingTablet) ButtonDown(key int) error {
	return sendBtnEvent(vg.deviceFile, []int{key}, btnStatePressed)
}

func (vg vDrawingTablet) ButtonUp(key int) error {
	return sendBtnEvent(vg.deviceFile, []int{key}, btnStateReleased)
}

func (vg vDrawingTablet) Pressure(Pressure float32) error {
	err := vg.sendInputEvent(inputEvent{
		Type:  evAbs,
		Code:  absPressure,
		Value: denormalizeInput(Pressure),
	})
	if err != nil {
		return err
	}
	return vg.syncEvents()
}

func (vg vDrawingTablet) Move(x, y float32) error {
	values := map[uint16]float32{}
	values[absX] = x
	values[absY] = y
	return vg.sendAbsEvent(values)
}

func (vg vDrawingTablet) sendAbsEvent(values map[uint16]float32) error {
	for code, value := range values {
		err := vg.sendInputEvent(inputEvent{
			Type:  evAbs,
			Code:  code,
			Value: denormalizeInput(value),
		})
		if err != nil {
			return err
		}
	}
	return vg.syncEvents()
}

func (vg vDrawingTablet) sendInputEvent(ev inputEvent) error {
	buf, err := inputEventToBuffer(ev)
	if err != nil {
		return fmt.Errorf("writing abs stick event failed: %v", err)
	}

	_, err = vg.deviceFile.Write(buf)
	if err != nil {
		return fmt.Errorf("failed to write abs stick event to device file: %v", err)
	}
	return vg.syncEvents()
}

func (vg vDrawingTablet) syncEvents() error {
	buf, err := inputEventToBuffer(inputEvent{
		Time:  syscall.Timeval{Sec: 0, Usec: 0},
		Type:  evSyn,
		Code:  uint16(synReport),
		Value: 0})
	if err != nil {
		return fmt.Errorf("writing sync event failed: %v", err)
	}
	_, err = vg.deviceFile.Write(buf)
	return err
}

func (vg vDrawingTablet) Close() error {
	return closeDevice(vg.deviceFile)
}

func createVDrawingTabletDevice(path string, name []byte) (fd *os.File, err error) {
	// This array is needed to register the event keys for the DrawingTablet device.
	keys := []uint16{
		ButtonToolPen,
		ButtonToolRubber,
		ButtonTouch,
		ButtonStylus,
		ButtonToolFinger,
		ButtonToolDoubletap,
		ButtonToolTripletap,
		ButtonToolQuadtap,
		ButtonToolQuinttap,
	}

	// This array is for the absolute events for the DrawingTablet device.
	abs_events := []uint16{
		absX,
		absY,
		absPressure,
		AbsTiltX,
		AbsTiltY,
	}

	deviceFile, err := createDeviceFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual DrawingTablet device: %v", err)
	}

	// register sync
	err = registerDevice(deviceFile, uintptr(evSyn))
	if err != nil {
		deviceFile.Close()
		return nil, fmt.Errorf("failed to register sync: %v", err)
	}

	// register button events
	err = registerDevice(deviceFile, uintptr(evKey))
	if err != nil {
		deviceFile.Close()
		return nil, fmt.Errorf("failed to register virtual DrawingTablet device: %v", err)
	}

	for _, code := range keys {
		err = ioctl(deviceFile, uiSetKeyBit, uintptr(code))
		if err != nil {
			deviceFile.Close()
			return nil, fmt.Errorf("failed to register key number %d: %v", code, err)
		}
	}

	// times
	err = registerDevice(deviceFile, uintptr(evMsc))
	if err != nil {
		deviceFile.Close()
		return nil, fmt.Errorf("failed to register sync: %v", err)
	}

	err = registerDevice(deviceFile, uintptr(MscTimestamp))
	if err != nil {
		deviceFile.Close()
		return nil, fmt.Errorf("failed to register sync: %v", err)
	}

	// register absolute events
	err = registerDevice(deviceFile, uintptr(evAbs))
	if err != nil {
		deviceFile.Close()
		return nil, fmt.Errorf("failed to register absolute event input device: %v", err)
	}

	for _, event := range abs_events {
		err = ioctl(deviceFile, uiSetAbsBit, uintptr(event))
		if err != nil {
			deviceFile.Close()
			return nil, fmt.Errorf("failed to register absolute event %v: %v", event, err)
		}
	}

	var absInfo = absInfoDrawingTablett()

	var absMin [absSize]int32
	var absMax [absSize]int32
	var absFuzz [absSize]int32
	var absFlat [absSize]int32

	for ec, a := range absInfo {
		absMax[ec] = a.Maximum
		absMin[ec] = a.Minimum
		absFuzz[ec] = a.Fuzz
		absFlat[ec] = a.Flat
	}

	err = setAbsSetup(deviceFile, absInfo)
	if err != nil {
		return nil, err
	}

	return createUsbDevice(deviceFile,
		uinputUserDev{
			Name: toUinputName(name),
			ID: inputID{
				Bustype: busUsb,
				Vendor:  0x4711,
				Product: 0x0818,
				Version: 1,
			},
			Absmin:  absMin,
			Absmax:  absMax,
			Absfuzz: absFuzz,
			Absflat: absFlat,
		})
}

func absInfoDrawingTablett() map[uint16]absInfo {
	return map[uint16]absInfo{
		absX:        {Maximum: MaximumAxisValue, Resolution: 12},
		absY:        {Maximum: MaximumAxisValue, Resolution: 12},
		absPressure: {Maximum: MaximumAxisValue, Resolution: 12},

		AbsTiltX: {Minimum: -90, Maximum: 90, Resolution: 12},
		AbsTiltY: {Minimum: -90, Maximum: 90, Resolution: 12},

		AbsMtSlot:        {Maximum: 4},
		AbsMtTrackingId:  {Maximum: 4},
		AbsMtTouchMajor:  {Maximum: MaximumAxisValue, Resolution: 12},
		AbsMtTouchMinor:  {Maximum: MaximumAxisValue, Resolution: 12},
		AbsMtOrientation: {Maximum: 1},
	}
}
