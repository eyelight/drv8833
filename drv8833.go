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
	"time"
)

// Device is a pair of motors without PWM
type Device struct {
	sleep, a1, a2, b1, b2 machine.Pin
}

// New returns a new DRV8833 driver
func New(sleep, a1, a2, b1, b2 machine.Pin) Device {
	return Device{
		sleep: sleep,
		a1:    a1,
		a2:    a2,
		b1:    b1,
		b2:    b2,
	}
}

// Configure configures the Device's Pins and sets the motors to sleep
func (d *Device) Configure() {
	d.sleep.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.Sleep()
	d.a1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.a2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.b1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.b2.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

// Sleep pulls the sleep pin to 'low'
func (d *Device) Sleep() {
	d.sleep.Low()
}

// Wake pulls the sleep pin 'high'
func (d *Device) Wake() {
	d.sleep.High()
}

// PWM is the interface necessary for controlling the motor
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	SetInverting(channel uint8, inverting bool)
}

// PWMDevice is a pair of motors with speed control
type PWMDevice struct {
	sleep machine.Pin
	a1    machine.Pin
	a2    machine.Pin
	b1    machine.Pin
	b2    machine.Pin
	a1ch  uint8
	a2ch  uint8
	b1ch  uint8
	b2ch  uint8
	pwm   PWM
}

// NewWithSpeed returns a new driver with PWM control
func NewWithSpeed(sleep, a1, a2, b1, b2 machine.Pin, pwm PWM) PWMDevice {
	return PWMDevice{
		sleep: sleep,
		a1:    a1,
		a2:    a2,
		b1:    b1,
		b2:    b2,
		pwm:   pwm,
	}
}

// Configure configures the PWMDevice. The pins,
// PWM interface, and channels must already be configured.
func (d *PWMDevice) Configure() {
	d.sleep.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.Sleep()
	d.a1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.a1ch, _ = d.pwm.Channel(d.a1)
	d.a2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.a2ch, _ = d.pwm.Channel(d.a2)
	d.b1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.b1ch, _ = d.pwm.Channel(d.b1)
	d.b2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.b2ch, _ = d.pwm.Channel(d.b2)
}

// Pulse turns motor on for a duration;
// the direction is whichever channel is passed as `ch1`
// slowDecay=false will make the non-PWM channel low, causing fast decay & vice versa
func (d *PWMDevice) Pulse(ch1, ch2 uint8, period, duration time.Duration, slowDecay bool) {
	err := d.pwm.SetPeriod(uint64(period))
	if err != nil {
		println("Pulse() error: " + err.Error())
	}
	d.pwm.Set(ch1, d.pwm.Top()/2) // half duty cycle for our "positive/pwm" pin (TODO: parameterize this?)
	if slowDecay == true {
		d.pwm.Set(ch2, d.pwm.Top()) // set opposite polarity to High for slow decay; see truth table above or datasheet
	} else {
		d.pwm.Set(ch2, 0) // set opposite polarity to Low for fast decay; see truth table above or datasheet
	}
	d.Wake()
	time.Sleep(duration)
	d.Sleep()
}

// Sleep pulls the sleep pin low
func (d *PWMDevice) Sleep() {
	d.sleep.Low()
}

// Wake pulls the sleep pin high
func (d *PWMDevice) Wake() {
	d.sleep.High()
}
