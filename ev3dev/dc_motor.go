// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

import (
	"fmt"
	"strings"
	"time"
)

var _ idSetter = (*DCMotor)(nil)

// DCMotor represents a handle to a dc-motor.
type DCMotor struct {
	id int

	err error
}

// Path returns the dc-motor sysfs path.
func (*DCMotor) Path() string { return DCMotorPath }

// Type returns "motor".
func (*DCMotor) Type() string { return motorPrefix }

// String satisfies the fmt.Stringer interface.
func (m *DCMotor) String() string {
	if m == nil {
		return motorPrefix + "*"
	}
	return fmt.Sprint(motorPrefix, m.id)
}

// Err returns the error state of the DCMotor and clears it.
func (m *DCMotor) Err() error {
	err := m.err
	m.err = nil
	return err
}

// setID satisfies the idSetter interface.
func (m *DCMotor) setID(id int) {
	*m = DCMotor{id: id}
}

// DCMotorFor returns a DCMotor for the given ev3 port name and driver. If the
// motor driver does not match the driver string, a DCMotor for the port is
// returned with a DriverMismatch error.
// If port is empty, the first dc-motor satisfying the driver name is returned.
func DCMotorFor(port, driver string) (*DCMotor, error) {
	id, err := deviceIDFor(port, driver, (*DCMotor)(nil))
	if id == -1 {
		return nil, err
	}
	return &DCMotor{id: id}, err
}

// Commands returns the available commands for the DCMotor.
func (m *DCMotor) Commands() ([]string, error) {
	return stringSliceFrom(attributeOf(m, commands))
}

// Command issues a command to the DCMotor.
func (m *DCMotor) Command(comm string) *DCMotor {
	if m.err != nil {
		return m
	}
	avail, err := m.Commands()
	if err != nil {
		m.err = err
		return m
	}
	ok := false
	for _, c := range avail {
		if c == comm {
			ok = true
			break
		}
	}
	if !ok {
		m.err = fmt.Errorf("ev3dev: command %q not available for %s (available:%q)", comm, m, avail)
		return m
	}
	m.err = setAttributeOf(m, command, comm)
	return m
}

// DutyCycle returns the current duty cycle value for the DCMotor.
func (m *DCMotor) DutyCycle() (int, error) {
	return intFrom(attributeOf(m, dutyCycle))
}

// DutyCycleSetpoint returns the current duty cycle setpoint value for the DCMotor.
func (m *DCMotor) DutyCycleSetpoint() (int, error) {
	return intFrom(attributeOf(m, dutyCycleSetpoint))
}

// SetDutyCycleSetpoint sets the duty cycle setpoint value for the DCMotor
func (m *DCMotor) SetDutyCycleSetpoint(sp int) *DCMotor {
	if m.err != nil {
		return m
	}
	if sp < -100 || sp > 100 {
		m.err = fmt.Errorf("ev3dev: invalid duty cycle setpoint: %d (valid -100 - 100)", sp)
		return m
	}
	m.err = setAttributeOf(m, dutyCycleSetpoint, fmt.Sprint(sp))
	return m
}

// Polarity returns the current polarity of the DCMotor.
func (m *DCMotor) Polarity() (Polarity, error) {
	p, err := stringFrom(attributeOf(m, polarity))
	return Polarity(p), err
}

// SetPolarity sets the polarity of the DCMotor
func (m *DCMotor) SetPolarity(p Polarity) *DCMotor {
	if m.err != nil {
		return m
	}
	if p != Normal && p != Inversed {
		m.err = fmt.Errorf("ev3dev: invalid polarity: %q (valid \"normal\" or \"inversed\")", p)
		return m
	}
	m.err = setAttributeOf(m, polarity, string(p))
	return m
}

// RampUpSetpoint returns the current ramp up setpoint value for the DCMotor.
func (m *DCMotor) RampUpSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampUpSetpoint))
}

// SetRampUpSetpoint sets the ramp up setpoint value for the DCMotor.
func (m *DCMotor) SetRampUpSetpoint(sp time.Duration) *DCMotor {
	if m.err != nil {
		return m
	}
	if sp < 0 || sp > 10000 {
		m.err = fmt.Errorf("ev3dev: invalid ramp up setpoint: %v (must be positive)", sp)
		return m
	}
	m.err = setAttributeOf(m, rampUpSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// RampDownSetpoint returns the current ramp down setpoint value for the DCMotor.
func (m *DCMotor) RampDownSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, rampDownSetpoint))
}

// SetRampDownSetpoint sets the ramp down setpoint value for the DCMotor.
func (m *DCMotor) SetRampDownSetpoint(sp time.Duration) *DCMotor {
	if m.err != nil {
		return m
	}
	if sp < 0 || sp > 10000 {
		m.err = fmt.Errorf("ev3dev: invalid ramp down setpoint: %v (must be positive)", sp)
		return m
	}
	m.err = setAttributeOf(m, rampDownSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// State returns the current state of the DCMotor.
func (m *DCMotor) State() (MotorState, error) {
	if m.err != nil {
		return 0, m.Err()
	}
	data, _, err := attributeOf(m, state)
	if err != nil {
		return 0, err
	}
	var stat MotorState
	for _, s := range strings.Split(data, " ") {
		bit, ok := motorStateTable[s]
		if !ok {
			return 0, fmt.Errorf("ev3dev: unrecognized motor state value: %s in [%s]", s, data)
		}
		stat |= bit
	}
	return stat, nil
}

// StopAction returns the stop action used when a stop command is issued
// to the DCMotor.
func (m *DCMotor) StopAction() (string, error) {
	return stringFrom(attributeOf(m, stopAction))
}

// SetStopAction sets the stop action to be used when a stop command is
// issued to the DCMotor.
func (m *DCMotor) SetStopAction(action string) *DCMotor {
	if m.err != nil {
		return m
	}
	avail, err := m.StopActions()
	if err != nil {
		m.err = err
		return m
	}
	ok := false
	for _, a := range avail {
		if a == action {
			ok = true
			break
		}
	}
	if !ok {
		m.err = fmt.Errorf("ev3dev: stop action %q not available for %s (available:%q)", action, m, avail)
		return m
	}
	m.err = setAttributeOf(m, stopAction, action)
	return m
}

// StopActions returns the available stop actions for the DCMotor.
func (m *DCMotor) StopActions() ([]string, error) {
	return stringSliceFrom(attributeOf(m, stopActions))
}

// TimeSetpoint returns the current time setpoint value for the DCMotor.
func (m *DCMotor) TimeSetpoint() (time.Duration, error) {
	return durationFrom(attributeOf(m, timeSetpoint))
}

// SetTimeSetpoint sets the time setpoint value for the DCMotor.
func (m *DCMotor) SetTimeSetpoint(sp time.Duration) *DCMotor {
	if m.err != nil {
		return m
	}
	m.err = setAttributeOf(m, timeSetpoint, fmt.Sprint(int(sp/time.Millisecond)))
	return m
}

// Uevent returns the current uevent state for the DCMotor.
func (m *DCMotor) Uevent() (map[string]string, error) {
	return ueventFrom(attributeOf(m, uevent))
}
