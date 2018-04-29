// Package nfsmon provides simple monitoring for nfs mounts. When a stale mount
// is detected, the remountFunc is called.
package nfsmon

import (
	"context"
	"strings"
	"sync"
	"syscall"
	"time"
)

type (
	// Mount defines a mount to watch.
	Mount struct {
		Server     string // Server defines the server mounted from.
		ServerPath string // ServerPath defines the location of the mount on the server.
		DestPath   string // DestPath defines the path to mount/monitor.
		MountOpts  string // MountOpts defines the mount options to use when mounting.
	}
)

var (
	// remountFunc gets called when a mount is detected as having gone stale.
	remountFunc    func(m Mount) error
	remountFuncTex = &sync.RWMutex{}

	// errCondition is checked by Watch. If true, Watch calls remountFunc. Default returns
	// true if stale nfs mount detected.
	errCondition    func(error) bool
	errConditionTex = &sync.RWMutex{}

	// mounts define which mounts to watch.
	mounts = []Mount{}
	mTex   = &sync.RWMutex{}
)

// WatchMount adds a mount to the list of mounts to watch.
func WatchMount(m Mount) {
	mTex.Lock()
	defer mTex.Unlock()
	for i := range mounts {
		if mounts[i].DestPath == m.DestPath {
			mounts[i] = m
			return
		}
	}
	mounts = append(mounts, m)
}

// UnwatchMount removes a mount from the list of mounts to watch.
func UnwatchMount(m Mount) {
	mTex.Lock()
	defer mTex.Unlock()
	for i := range mounts {
		if mounts[i].DestPath == m.DestPath {
			mounts = append(mounts[:i], mounts[i+1:]...)
			return
		}
	}
}

// SetErrConditionFunc sets the errCondition function in a threadsafe manner.
// The "errCondition" function is checked by Watch. If true, it calls remountFunc.
// Default returns true if stale nfs mount detected.
func SetErrConditionFunc(f func(error) bool) {
	errConditionTex.Lock()
	defer errConditionTex.Unlock()
	errCondition = f
}

// getErrConditionFunc gets the configured or default error condition function.
func getErrConditionFunc() func(error) bool {
	errConditionTex.RLock()
	defer errConditionTex.RUnlock()
	if errCondition == nil {
		return errConditionFunc
	}
	return errCondition
}

// SetRemountFunc sets the remount function in a threadsafe manner.
func SetRemountFunc(f func(Mount) error) {
	remountFuncTex.Lock()
	defer remountFuncTex.Unlock()
	remountFunc = f
}

// getRemountFunc gets the configured remount function.
func getRemountFunc() func(Mount) error {
	remountFuncTex.RLock()
	defer remountFuncTex.RUnlock()
	return remountFunc
}

// errConditionFunc is the default error condition to check for; a stale nfs mount.
func errConditionFunc(err error) bool {
	return err == syscall.ESTALE && strings.Contains(err.Error(), "NFS")
}

// WatchCfg allows for configuring the watch function.
type WatchCfg struct {
	// NumRetries defines how many times to retry mounting.
	NumRetries int
	// WatchFreq defines how frequently to check mount destination path.
	WatchFreq time.Duration
	// todo: add RemountBackoff - time to sleep before retrying the mount
}

// Watch watches the configured mountsand calls remountFunc if errCondition is true.
// Configured via functional options. For default config, run with Watch(ctx).
func Watch(ctx context.Context, opts ...func(*WatchCfg)) {
	cfg := WatchCfg{
		NumRetries: 3,
		WatchFreq:  time.Second * 30,
	}

	// set config options (functional parameters)
	for i := range opts {
		opts[i](&cfg)
	}

	for {
		select {
		case <-time.Tick(cfg.WatchFreq):
			mTex.RLock()
			tMounts := mounts
			mTex.RUnlock()

			for i := range tMounts {
				err := syscall.Statfs(tMounts[i].DestPath, nil)
				if err != nil {
					if getErrConditionFunc()(err) {
						retry := 0
					remount:
						if getRemountFunc() == nil {
							continue
						}
						// consumers can add timeout in remount func
						err := getRemountFunc()(tMounts[i])
						if err != nil {
							retry++
							if retry < cfg.NumRetries {
								time.Sleep(time.Second)
								goto remount
							}
							continue
						}
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
