package state

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type WriteFile struct {
	Filename string
	Data     []byte
	Perm     os.FileMode // default 0644
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

func (s WriteFile) Apply() (RollbackCleaner, error) {
	if s.Perm == 0 {
		s.Perm = 0644
	}
	rc, err := newWriteFileRC(s.Filename)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(s.Filename, s.Data, s.Perm); err != nil {
		rc.Clean()
		return nil, err
	}
	return rc, nil
}

func (s WriteFile) String() string {
	perm := os.FileMode(0644)
	if s.Perm != 0 {
		perm = s.Perm
	}
	return fmt.Sprintf("WriteFile(%s, %3o)", s.Filename, perm)
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

type MakeDir struct {
	Path string
	Perm os.FileMode // default 0755
}

type makeDirRC struct {
	Path string
}

func (rc makeDirRC) Rollback() error {
	if rc.Path == "" {
		return nil
	}
	return os.Remove(rc.Path)
}

func (rc makeDirRC) Clean() error {
	return nil
}

func (s MakeDir) Apply() (RollbackCleaner, error) {
	if s.Perm == 0 {
		s.Perm = 0755
	}
	if stat, err := os.Stat(s.Path); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "unknown error")
		}
	} else {
		if !stat.IsDir() {
			return nil, fmt.Errorf("%s exists and is not a directory", s.Path)
		}
		return &dummyRC{}, nil
	}

	if err := os.MkdirAll(s.Path, s.Perm); err != nil {
		return nil, errors.Wrap(err, "fail to make dir")
	}

	return makeDirRC{Path: s.Path}, nil
}

func (s MakeDir) String() string {
	perm := os.FileMode(0755)
	if s.Perm != 0 {
		perm = s.Perm
	}
	return fmt.Sprintf("MakeDir(%s, %3o)", s.Path, perm)
}

type Symlink struct {
	Src  string
	Link string
}

type symlinkRC struct {
	OldSrc string
	Link   string
}

func (rc symlinkRC) Rollback() error {
	if rc.OldSrc == "" {
		return nil
	}
	return os.Symlink(rc.OldSrc, rc.Link)
}

func (rc symlinkRC) Clean() error { return nil }

func (s Symlink) Apply() (RollbackCleaner, error) {
	if _, err := os.Stat(s.Src); err != nil {
		if os.IsNotExist(err) {
			return dummyRC{}, nil
		}
		return nil, err
	}
	var oldSrc string
	if stat, err := os.Lstat(s.Link); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		if stat.Mode()&os.ModeSymlink == 0 {
			return nil, fmt.Errorf("%s is not a symlink", s.Link)
		} else {
			var err error
			oldSrc, err = os.Readlink(s.Link)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to find old source of link %s", s.Link)
			}
		}
	}
	if oldSrc != "" {
		if err := os.Remove(s.Link); err != nil {
			return nil, err
		}
	}
	if err := os.Symlink(s.Src, s.Link); err != nil {
		return nil, err
	}
	return symlinkRC{OldSrc: oldSrc, Link: s.Link}, nil
}

func (s Symlink) String() string {
	return fmt.Sprintf("Symlink(%s, %s)", s.Src, s.Link)
}
