// Package nfsmon provides simple monitoring for nfs mounts. When a stale mount
// is detected, the RemountFunc is called.
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
	// RemountFuc gets called when a mount is detected as having gone stale.
	RemountFunc func(m Mount) error
	// WatchFreq defines how frequently to check mount destination path.
	WatchFreq = time.Second * 30
	// mounts define which mounts to watch.
	mounts = []Mount{}
	// mTex	is the mounts' mutual exclusive lock.
	mTex = sync.RWMutex{}
)

// WatchMount adds a mount to the list of mounts to watch.
// func WatchMount(m Remounter) {
func WatchMount(m Mount) {
	mounts = append(mounts, m)
}

// UnwatchMount removes a mount from the list of mounts to watch.
func UnwatchMount(m Mount) {
	for i := range mounts {
		if mounts[i].DestPath == m.DestPath {
			mounts = append(mounts[:i], mounts[i+1:]...)
			return
		}
	}
}

// ErrCondition is checked by Watch. If true, Watch calls RemountFunc. Default returns
// true if stale nfs mount detected.
var ErrCondition func(error) bool = errConditionFunc

// errConditionFunc is the default error condition to check for; a stale nfs mount.
func errConditionFunc(err error) bool {
	return err == syscall.ESTALE && strings.Contains(err.Error(), "NFS")
}

// Watch watches the configured mounts and calls RemountFunc if ErrCondition is true.
func Watch(ctx context.Context) {
	for {
		select {
		case <-time.Tick(WatchFreq):
			mTex.RLock()
			tMounts := mounts
			mTex.RUnlock()

			for i := range tMounts {
				err := syscall.Statfs(tMounts[i].DestPath, nil)
				if err != nil {
					if ErrCondition(err) {
						retry := 0
					remount:
						// todo: add timeout
						err := RemountFunc(tMounts[i])
						if err != nil {
							retry++
							if retry < 3 {
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
