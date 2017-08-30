package state

type NginxReload struct{}

func (NginxReload) Apply() (Unapplyer, error) {
	if err := execSplitCmd("nginx -t"); err != nil {
		return nil, err
	}
	return dummyU{}, execSplitCmd("nginx -s reload")
}

func (NginxReload) String() string { return "NginxReload()" }
