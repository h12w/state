package state

type NginxReload struct{}

func (NginxReload) Apply() (RollbackCleaner, error) {
	if err := execSplitCmd("nginx -t"); err != nil {
		return nil, err
	}
	return dummyRC{}, execSplitCmd("nginx -s reload")
}

func (NginxReload) String() string { return "NginxReload()" }
