// Package drv8833 provides a driver for the DRV8833 dual h-bridge chip
// able to drive DC motors, bipolar steppers, solenoids, and other inductive loads
// The DRV8833 has a wide power supply range from 2.7v - 10.8v
// Included are methods that seem appropriate for DC motors Run() & latching solenoids Pulse(), but not steppers
//
// Create a PWM-aware PWMDevice or non-PWM aware Device
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

// Device is a pair of h-bridges without PWM
type Device struct {
	sleep, a1pin, a2pin, b1pin, b2pin machine.Pin
}

// New returns a new Device (non-PWM)
func New(sleep, a1pin, a2pin, b1pin, b2pin machine.Pin) Device {
	return Device{
		sleep: sleep,
		a1pin: a1pin,
		a2pin: a2pin,
		b1pin: b1pin,
		b2pin: b2pin,
	}
}

// Configure configures the Device pins and sets the h-bridges to 'sleep'
func (d *Device) Configure() {
	d.sleep.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.Sleep() // no funny business before we want to use these
	d.a1pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.a2pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.b1pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.b2pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

// Sleep pulls the sleep pin low
func (d *Device) Sleep() {
	d.sleep.Low()
}

// Wake pulls the sleep pin high
func (d *Device) Wake() {
	d.sleep.High()
}

// BrakeA pulls both A1 & A2 pins high
func (d *Device) BrakeA() {
	d.a1pin.High()
	d.a2pin.High()
}

// BrakeB pulls both B1 & B2 pins high
func (d *Device) BrakeB() {
	d.b1pin.High()
	d.b2pin.High()
}

// CoastA pulls both A1 & A2 pins low
func (d *Device) CoastA() {
	d.a1pin.Low()
	d.a2pin.Low()
}

// CoastB pulls both B1 & B2 pins low
func (d *Device) CoastB() {
	d.b1pin.Low()
	d.b2pin.Low()
}

// PWM is an interface for interacting with a machine.pwmGroup
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Get(channel uint8) (value uint32)
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	SetInverting(channel uint8, inverting bool)
}

// PWMDevice is a pair of h-bridges as found on the DRV8833 chip, with PWM support
type PWMDevice struct {
	sleep machine.Pin // PWM not necessary
	a1pin machine.Pin // must be PWM pin
	a2pin machine.Pin // must be PWM pin
	b1pin machine.Pin // must be PWM pin
	b2pin machine.Pin // must be PWM pin
	A1    uint8       // PWM channel used for a1
	A2    uint8       // PWM channel used for a2
	B1    uint8       // PWM channel used for b1
	B2    uint8       // PWM channel used for b2
	PwmA  PWM         // the PWM used by h-bridge A
	PwmB  PWM         // the PWM used by h-bridge B
}

// NewWithSpeed configures two PWMs and returns a new PWMDevice given some pins and a configured PWMConfig
func NewWithSpeed(sleep, a1pin, a2pin, b1pin, b2pin machine.Pin, pwmA, pwmB PWM, pwmConfA, pwmConfB machine.PWMConfig) PWMDevice {
	err := pwmA.Configure(pwmConfA)
	if err != nil {
		println("error Configuring DRV8833 pwmA: " + err.Error())
	}
	err = pwmB.Configure(pwmConfB)
	if err != nil {
		println("error Configuring DRV8833 pwmB: " + err.Error())
	}
	return PWMDevice{
		sleep: sleep,
		a1pin: a1pin,
		a2pin: a2pin,
		b1pin: b1pin,
		b2pin: b2pin,
		A1:    0,
		A2:    0,
		B1:    0,
		B2:    0,
		PwmA:  pwmA,
		PwmB:  pwmB,
	}
}

// Configure configures the PWMDevice, setting Pins as outputs,
// and pwm channels obtained;
// both h-bridges will begin in 'sleep' mode
func (d *PWMDevice) Configure() {
	d.sleep.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.Sleep() // no funny business before we want to use these

	// Configure pins as output & obtain PWM channels
	d.a1pin.Configure(machine.PinConfig{Mode: machine.PinPWM})
	a1, err := d.PwmA.Channel(d.a1pin)
	if err != nil {
		println("error obtaining DRV8833 a1 channel: " + err.Error())
	}
	d.A1 = a1

	d.a2pin.Configure(machine.PinConfig{Mode: machine.PinPWM})
	a2, err := d.PwmA.Channel(d.a2pin)
	if err != nil {
		println("error obtaining DRV8833 a2 channel: " + err.Error())
	}
	d.A2 = a2

	d.b1pin.Configure(machine.PinConfig{Mode: machine.PinPWM})
	b1, err := d.PwmB.Channel(d.b1pin)
	if err != nil {
		println("error obtaining DRV8833 b1 channel: " + err.Error())
	}
	d.B1 = b1

	d.b2pin.Configure(machine.PinConfig{Mode: machine.PinPWM})
	b2, err := d.PwmB.Channel(d.b2pin)
	if err != nil {
		println("error obtaining DRV8833 b2 channel: " + err.Error())
	}
	d.B2 = b2
}

// RunA will energize the load connected to h-bridge A at a given duty %;
// polarity is chosen by whichever channel is passed as 'ch1';
// for more about slow/fast decay, see DR8833 datasheet Section 7.3.2
func (d *PWMDevice) RunA(duty, ch1, ch2 uint8, slowDecay bool) {
	if duty > 100 {
		duty = 100
	}
	d.PwmA.Set(ch1, d.PwmA.Top()*uint32(duty)/100)
	if slowDecay == true {
		d.PwmA.Set(ch2, d.PwmA.Top())
	} else {
		d.PwmA.Set(ch2, 0)
	}
	if d.sleep.Get() == false {
		d.Wake()
	}
}

// RunB will energize the load connected to h-bridge B at a given duty %;
// polarity is chosen by whichever channel is passed as 'ch1';
// for more about slow/fast decay, see DR8833 datasheet Section 7.3.2
func (d *PWMDevice) RunB(duty, ch1, ch2 uint8, slowDecay bool) {
	if duty > 100 {
		duty = 100
	}
	d.PwmB.Set(ch1, d.PwmB.Top()*uint32(duty)/100)
	if slowDecay == true {
		d.PwmB.Set(ch2, d.PwmB.Top())
	} else {
		d.PwmB.Set(ch2, 0)
	}
	if d.sleep.Get() == false {
		d.Wake()
	}
}

// PulseA will (blockingly) pulse the load connected to h-bridge A at a given duty %, for a given duration;
// polarity is chosen by whichever channel is passed as 'ch1';
// the DRV8833 will be put into sleep mode after the pulse duration;
// for more about slow/fast decay, see DRV8833 datasheet Section 7.3.2
func (d *PWMDevice) PulseA(duty, ch1, ch2 uint8, dur time.Duration, slowDecay bool) {
	if duty > 100 {
		duty = 100
	}
	switch slowDecay {
	case true:
		defer d.PwmA.Set(ch1, 0)
		defer d.PwmA.Set(ch2, 0)
		d.PwmA.Set(ch1, d.PwmA.Top())
		d.PwmA.Set(ch2, d.PwmA.Top()*uint32(duty)/100)
	case false:
		defer d.PwmA.Set(ch1, 0)
		d.PwmA.Set(ch1, d.PwmA.Top()*uint32(duty)/100)
		d.PwmA.Set(ch2, 0)
	}
	d.Wake()
	time.Sleep(dur)
	d.Sleep()
}

// PulseB will (blockingly) pulse the load connected to h-bridge B for a given duration;
// polarity is chosen by whichever channel is passed as 'ch1';
// the DRV8833 will be put into sleep mode after the pulse duration;
// for more about slow/fast decay DRV8833 datasheet Section 7.3.2
func (d *PWMDevice) PulseB(duty, ch1, ch2 uint8, dur time.Duration, slowDecay bool) {
	if duty > 100 {
		duty = 100
	}
	switch slowDecay {
	case true:
		defer d.PwmB.Set(ch1, 0)
		defer d.PwmB.Set(ch2, 0)
		d.PwmB.Set(ch1, d.PwmB.Top())
		d.PwmB.Set(ch2, d.PwmB.Top()*uint32(duty)/100)
	case false:
		defer d.PwmB.Set(ch1, 0)
		d.PwmB.Set(ch1, d.PwmB.Top()*uint32(duty)/100)
		d.PwmB.Set(ch2, 0) // to use "slow decay", pass 'd.pwmA.Top()' instead of '0' as the second arg to Set
	}
	d.Wake()
	time.Sleep(dur)
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

// BrakeA pulls both PWM channels high for h-bridge A
func (d *PWMDevice) BrakeA() {
	d.PwmA.Set(d.A1, d.PwmA.Top())
	d.PwmA.Set(d.A2, d.PwmA.Top())
}

// BrakeB pulls both PWM channels high for h-bridge B
func (d *PWMDevice) BrakeB() {
	d.PwmB.Set(d.B1, d.PwmB.Top())
	d.PwmB.Set(d.B2, d.PwmB.Top())
}

// CoastA pulls both PWM channels low for h-bridge A
func (d *PWMDevice) CoastA() {
	d.PwmA.Set(d.A1, 0)
	d.PwmA.Set(d.A2, 0)
}

// CoastB pulls both PWM channels low for h-bridge B
func (d *PWMDevice) CoastB() {
	d.PwmB.Set(d.B1, 0)
	d.PwmB.Set(d.B2, 0)
}
