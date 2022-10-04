package auto

import (
	carmodel "github.com/slavsan/godiss/examples/cars"
)

type Factory struct {
	Name string
}

type Mechanic struct {
	Skills     []string
	Colleagues []*Mechanic
}

type Manager struct {
	Pointer *Mechanic
}

type tool struct {
	name string
}

type IMechanic interface {
	DoWork()
	BuildCamaro() (*carmodel.Camaro, error)
}
