package state

import "log"

type (
	State interface {
		Apply() (RollbackCleaner, error)
	}
	RollbackCleaner interface {
		Rollback() error
		Clean() error
	}
)

func ExecAll(execers ...State) error {
	rcs := make([]RollbackCleaner, 0, len(execers))
	for _, execer := range execers {
		rc, err := execer.Apply()
		if err != nil {
			for i := len(rcs) - 1; i >= 0; i-- {
				if err := rcs[i].Rollback(); err != nil {
					log.Print(err)
				}
			}
			return err
		}
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
