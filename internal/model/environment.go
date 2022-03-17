package model

import (
	"errors"
	"fmt"
)

// Environment specifies the execution environment of this app.
// Note that Environment satifies the Value interface by
// implementing String and Set. This enables its usage with Flag package,
// particularly flag.Var
type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

func (e *Environment) String() string {
	switch *e {
	case Production:
		return "production"
	case Development:
		return "development"
	default:
		return "development"
	}
}

func (e *Environment) Set(value string) error {
	switch value {
	case "production":
		*e = Production
		return nil
	case "development":
		*e = Development
		return nil
	default:
		return errors.New(fmt.Sprintf("unrecognised environment string %s", value))
	}
}
