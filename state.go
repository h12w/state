package state

import (
	"fmt"
	"log"
)

type (
	State interface {
		Apply() (RollbackCleaner, error)
	}
	RollbackCleaner interface {
		Rollback() error
		Clean() error
	}
)

func Apply(states ...State) error {
	rcs := make([]RollbackCleaner, 0, len(states))
	for _, state := range states {
		rc, err := state.Apply()
		if err != nil {
			for i := len(rcs) - 1; i >= 0; i-- {
				if err := rcs[i].Rollback(); err != nil {
					log.Print(err)
				}
			}
			return err
		}
		fmt.Println(state)
		rcs = append(rcs, rc)
	}
	for i := len(rcs) - 1; i >= 0; i-- {
		if err := rcs[i].Clean(); err != nil {
			log.Print(err)
		}
	}
	return nil
}

type dummyRC struct{}

func (rc dummyRC) Rollback() error { return nil }

func (rc dummyRC) Clean() error { return nil }
