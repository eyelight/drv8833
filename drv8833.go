// Package drv8833 provides a driver for the DRV8833 dual h-bridge chip
// able to drive DC motors, bipolar steppers, solenoids, and other inductive loads
// The DRV8833 has a wide power supply range from 2.7v - 10.8v
//
// Datasheet: https://www.ti.com/lit/ds/symlink/drv8833.pdf
//
// Pins
// IN1  | H-bridge 1
// IN2  | H-bridge 1
// IN3  | H-bridge 2
// IN4  | H-bridge 2
// GND  |
// VCC  | 3-10v
// ULT  | Low = Sleep; High = Run
// EEP  | Output protection
// OUT1 |
// OUT2 |
// OUT3 |
// OUT4 |
//
// Truth Table (per H-bridge)
// | IN1 | IN2 | OUT1 | OUT2 | FUNCTION |
// | --- | --- | ---- | ---- | -------- |
// |  0  |  0  |   Z  |   Z  | Coast / fast decay |
// |  0  |  1  |   L  |   H  | Reverse |
// |  1  |  0  |   H  |   L  | Forward |
// |  1  |  1  |   L  |   L  | Brake / slow decay |
//
// PWM Control Truth Table
// | IN1 | IN2 | Function |
// | --- | --- | -------- |
// | PWM |  0  | Forward PWM, fast decay |
// |  1  | PWM | Forward PWM, slow decay |
// |  0  | PWM | Reverse PWM, fast decay |
// | PWM |  1  | Reverse PWM, slow decay |
//
package drv8833

import (
	"machine"
)

// Device is a pair of motors without PWM
type Device struct {
}

// New returns a new DRV8833 driver
func New() Device {
	return Device{}
}

// Configure configures the Device
func (d *Device) Configure() {
}

// PWM is the interface necessary for controlling the motor
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
}

// PWMDevice is a pair of motors with speed control
type PWMDevice struct {
}

// NewWithSpeed returns a new driver with PWM control
func NewWithSpeed() PWMDevice {
	return PWMDevice{}
}

// Configure configures the PWMDevice. The pins,
// PWM interface, and channels must already be configured.
func (d *PWMDevice) Configure() (err error) {
	d.Stop()
	return
}

// Forward turns motor on in forward direction
