// Package drv8833 provides a driver for the DRV8833 dual h-bridge chip
// Datasheet: https://www.ti.com/lit/ds/symlink/drv8833.pdf
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
