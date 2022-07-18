package uinput

import (
	"testing"
	"time"
)

// This event inputs the konami keys to debug using evtest.
func TestGamepadKonamiKeys(t *testing.T) {
	vg, err := CreateGamepad("/dev/uinput", []byte("Hot gophers in your area"), 0xDEAD, 0xBEEF)
	if err != nil {
		t.Fatalf("Failed to create the virtual gamepad. Last error was: %s\n", err)
	}
	var keys = []int{
		ButtonDpadUp,
		ButtonDpadDown,
		ButtonDpadLeft,
		ButtonDpadRight,
		ButtonSouth,
		ButtonEast,
	}

	for i := range keys {
		err = vg.ButtonPress(keys[i])
		if err != nil {
			t.Fatalf("Failed to send button press. Last error was: %s\n", err)
		}
		sleepForaBit()
	}

}

func TestGamepadtRightKeys(t *testing.T) {
	vg, err := CreateGamepad("/dev/uinput", []byte("Hot gophers in your area"), 0xDEAD, 0xBEEF)
	if err != nil {
		t.Fatalf("Failed to create the virtual gamepad. Last error was: %s\n", err)
	}
	var keys = []int{
		ButtonSouth,
		ButtonEast,
		ButtonNorth,
		ButtonWest}

	for i := range keys {
		err = vg.ButtonPress(keys[i])
		if err != nil {
			t.Fatalf("Failed to send button press. Last error was: %s\n", err)
		}
		sleepForaBit()
	}
}

func TestGamepadtDirectionKeys(t *testing.T) {
	vg, err := CreateGamepad("/dev/uinput", []byte("Hot gophers in your area"), 0xDEAD, 0xBEEF)
	if err != nil {
		t.Fatalf("Failed to create the virtual gamepad. Last error was: %s\n", err)
	}
	var keys = []int{
		ButtonDpadUp,
		ButtonDpadDown,
		ButtonDpadLeft,
		ButtonDpadRight}

	for i := range keys {
		err = vg.ButtonPress(keys[i])
		if err != nil {
			t.Fatalf("Failed to send button press. Last error was: %s\n", err)
		}
		sleepForaBit()
	}
}

func TestAxisMovement(t *testing.T) {
	vg, err := CreateGamepad("/dev/uinput", []byte("Hot gophers in your area"), 0xDEAD, 0xBEEF)
	if err != nil {
		t.Fatalf("Failed to create the virtual gamepad. Last error was: %s\n", err)
	}

	err = vg.LeftStickMove(0.2, 1.0)
	if err != nil {
		t.Fatalf("Failed to send axis event. Last error was: %s\n", err)
	}

	err = vg.RightStickMove(0.2, 1.0)
	if err != nil {
		t.Fatalf("Failed to send axis event. Last error was: %s\n", err)
	}
}

func sleepForaBit() {
	time.Sleep(150 * time.Millisecond)
}
