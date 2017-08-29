package state

type NginxTest struct{}

func (e NginxTest) Exec() (RollbackCleaner, error) {
	return dummyRC{}, execCmd("nginx -t")
}

type NginxReload struct{}

func (e NginxReload) Exec() (RollbackCleaner, error) {
	return dummyRC{}, execCmd("nginx -s reload")
}
