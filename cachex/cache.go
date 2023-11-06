package cachex

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hilaily/kit/pathx"
	"github.com/sirupsen/logrus"
)

var (
	_ ICache[string] = &_cacheSrv[string]{}
)

// IStore ...
type ICache[T any] interface {
	Set(*T) error
	Get() (*T, error)
	Update(func(*T)) error

	MustSet(*T)
	MustGet() *T
	MustUpdate(func(*T))
}

// New ...
func New[T any](_filepath string) (*_cacheSrv[T], error) {
	c := &_cacheSrv[T]{
		file:     _filepath,
		lockFile: _filepath + ".lock",
		dir:      filepath.Dir(_filepath),
	}
	var t T
	if !pathx.IsExist(_filepath) {
		err := c.Set(&t)
		return c, err
	}
	return c, nil
}

type _cacheSrv[T any] struct {
	lockFile string
	file     string
	dir      string
}

func (rc *_cacheSrv[T]) MustSet(data *T) {
	err := rc.Set(data)
	if err != nil {
		logrus.Panicf(err.Error())
	}
}

func (rc *_cacheSrv[T]) MustGet() *T {
	r, err := rc.Get()
	if err != nil {
		logrus.Panicf(err.Error())
	}
	return r
}

func (rc *_cacheSrv[T]) MustUpdate(f func(*T)) {
	err := rc.Update(f)
	if err != nil {
		logrus.Panicf(err.Error())
	}
}

// Set ...
func (rc *_cacheSrv[T]) Set(data *T) (e error) {
	err := rc.tryLock()
	if err != nil {
		return err
	}
	defer func() {
		err := rc.unlock()
		e = err
	}()
	return rc.set(data)
}

func (rc *_cacheSrv[T]) set(data *T) (e error) {
	jobJSON, err := json.MarshalIndent(data, "  ", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(rc.file, jobJSON, 0777)
	if err != nil {
		return err
	}
	return nil
}

// Get ...
func (rc *_cacheSrv[T]) Get() (*T, error) {
	var val T
	jobJSON, err := os.ReadFile(rc.file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logrus.Warnf("%s is not exist", rc.file)
			return &val, nil
		}
		return &val, err
	}

	err = json.Unmarshal([]byte(jobJSON), &val)
	if err != nil {
		return &val, err
	}
	return &val, err
}

// Update ...
func (rc *_cacheSrv[T]) Update(f func(i *T)) (e error) {
	err := rc.tryLock()
	if err != nil {
		return err
	}
	defer func() {
		err := rc.unlock()
		e = err
	}()
	i, err := rc.Get()
	if err != nil {
		return err
	}
	f(i)
	err = rc.set(i)
	return err
}

func (rc *_cacheSrv[T]) tryLock() error {
	count := 10
	get := false
	for i := 1; i <= count; i++ {
		if pathx.IsExist(rc.lockFile) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		get = true
		break
	}
	if get {
		f, err := os.OpenFile(rc.lockFile, os.O_CREATE, 0777)
		if err != nil {
			return fmt.Errorf("write lock file fail %w", err)
		}
		defer f.Close()
		return nil
	}
	return fmt.Errorf("can not get lock, timeout")
}

func (rc *_cacheSrv[T]) unlock() error {
	return os.Remove(rc.lockFile)
}
