package state

import (
	"fmt"
	"log"
)

type (
	State interface {
		Apply() (Unapplyer, error)
	}
	Unapplyer interface {
		Unapply() error
	}
)

func Apply(states ...State) error {
	rcs := make([]Unapplyer, 0, len(states))
	for _, state := range states {
		rc, err := state.Apply()
		if err != nil {
			for i := len(rcs) - 1; i >= 0; i-- {
				if err := rcs[i].Unapply(); err != nil {
					log.Print(err)
				}
			}
			return err
		}
		fmt.Println(state)
		rcs = append(rcs, rc)
	}
	return nil
}

type dummyU struct{}

func (rc dummyU) Unapply() error { return nil }
