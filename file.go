package state

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type WriteFile struct {
	Filename string
	Data     []byte
	Perm     os.FileMode // default 0644
}

type writeFileU struct {
	Filename   string
	BackupData []byte
}

func (rc writeFileU) Unapply() error {
	if rc.BackupData == nil {
		return os.Remove(rc.Filename)
	}
	return ioutil.WriteFile(rc.Filename, rc.BackupData, 0644)
}

func newWriteFileU(filename string) (Unapplyer, error) {
	oriFile, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return writeFileU{Filename: filename}, nil
		}
		return nil, err
	}
	defer oriFile.Close()
	backupData, err := ioutil.ReadAll(oriFile)
	if err != nil {
		return nil, err
	}
	return writeFileU{
		Filename:   filename,
		BackupData: backupData,
	}, nil
}

func (s WriteFile) Apply() (Unapplyer, error) {
	if s.Perm == 0 {
		s.Perm = 0644
	}
	rc, err := newWriteFileU(s.Filename)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(s.Filename, s.Data, s.Perm); err != nil {
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

type MakeDir struct {
	Path string
	Perm os.FileMode // default 0755
}

type makeDirU struct {
	Path string
}

func (rc makeDirU) Unapply() error {
	if rc.Path == "" {
		return nil
	}
	return os.Remove(rc.Path)
}

func (s MakeDir) Apply() (Unapplyer, error) {
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
		return &dummyU{}, nil
	}

	if err := os.MkdirAll(s.Path, s.Perm); err != nil {
		return nil, errors.Wrap(err, "fail to make dir")
	}

	return makeDirU{Path: s.Path}, nil
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

type symlinkU struct {
	OldSrc string
	Link   string
}

func (rc symlinkU) Unapply() error {
	if rc.OldSrc == "" {
		return nil
	}
	src, err := os.Readlink(rc.Link)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if src == rc.OldSrc {
		return nil
	} else {
		os.Remove(rc.Link)
	}
	return os.Symlink(rc.OldSrc, rc.Link)
}

func (s Symlink) Apply() (Unapplyer, error) {
	if _, err := os.Stat(s.Src); err != nil {
		if os.IsNotExist(err) {
			return dummyU{}, nil
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
	return symlinkU{OldSrc: oldSrc, Link: s.Link}, nil
}

func (s Symlink) String() string {
	return fmt.Sprintf("Symlink(%s, %s)", s.Src, s.Link)
}
