package other

type Vehicle struct {
	Doors int
}

func (v *Vehicle) StartEngine() error {
	return nil
}

func (v Vehicle) StopEngine() error {
	return nil
}

func RateVehicle() int {
	return 0
}
