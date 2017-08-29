package state

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type WriteFile struct {
	Filename string
	Data     []byte
	Perm     os.FileMode
}

type writeFileRC struct {
	Filename       string
	BackupFilename string
}

func (rc writeFileRC) Rollback() error {
	if rc.BackupFilename == "" {
		return os.Remove(rc.Filename)
	}
	return os.Rename(rc.BackupFilename, rc.Filename)
}

func (rc writeFileRC) Clean() error {
	if rc.BackupFilename == "" {
		return nil
	}
	return os.Remove(rc.BackupFilename)
}

func newWriteFileRC(filename string) (RollbackCleaner, error) {
	oriFile, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return writeFileRC{Filename: filename}, nil
		}
		return nil, err
	}
	defer oriFile.Close()
	backupFilename, err := backupFile(oriFile)
	if err != nil {
		return nil, err
	}
	return writeFileRC{
		Filename:       filename,
		BackupFilename: backupFilename,
	}, nil
}

func (e WriteFile) Apply() (RollbackCleaner, error) {
	rc, err := newWriteFileRC(e.Filename)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(e.Filename, e.Data, e.Perm); err != nil {
		rc.Clean()
		return nil, err
	}
	return rc, nil
}

func backupFile(r io.Reader) (string, error) {
	backupFile, err := ioutil.TempFile(os.TempDir(), "state")
	if err != nil {
		return "", errors.Wrap(err, "fail to create backup file")
	}
	defer backupFile.Close()
	if _, err := io.Copy(backupFile, r); err != nil {
		return "", errors.Wrap(err, "fail to backup file")
	}
	return backupFile.Name(), nil
}
