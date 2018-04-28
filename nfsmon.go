package nfsmon

import (
	"context"
	"fmt"
	// "os"
	// "strings"
	// "syscall"
	"time"
)

type (
	// Mount defines a mount to watch
	Mount struct {
		Server     string
		ServerPath string
		DestPath   string
		MountOpts  string

		// RemountFunc func() error
	}

	// Remounter interface {
	// 	Remount() error
	// }

)

var RemountFunc func(m Mount) error
var WatchTime = time.Second * 30

// func NewDefaultMount()

// func (m Mount) Remount() error {
// 	return nil
// }

var mounts = []Mount{}

// var mounts = []Remounter{}

// WatchMount adds a mount to the list of mounts to watch.
// func WatchMount(m Remounter) {
func WatchMount(m Mount) {
	// mounts = append(mounts, Mount(m))
	// mounts = append(mounts, m.(Mount))
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

func Watch(ctx context.Context) {
	for {
		select {
		case <-time.Tick(WatchTime):
			// mTex.Lock()
			tMounts := mounts
			// mTex.Unlock()

			for i := range tMounts {
				err := RemountFunc(tMounts[i])
				if err != nil {
					fmt.Printf("ERR - %s\n", err.Error())
					continue
				}
			}
			// time.Sleep(time.Second * 30)
			// Watch()
		case <-ctx.Done():
			fmt.Println("Context done", ctx.Err())
			return
		}
	}
}

// func watch() {
// 	// path := "/tmp/thing"
// 	// watch:
// 	// mTex.Lock()
// 	tMounts := mounts
// 	// mTex.Unlock()

// 	buf := syscall.Statfs_t{}
// 	for i := range tMounts {
// 		err := syscall.Statfs(tMounts[i].DestPath, &buf)
// 		if err != nil {
// 			if err == syscall.ESTALE && strings.Contains(err.Error(), "NFS") {
// 				fmt.Printf("Stale file descriptor - %s\n", err.Error())
// 				// todo: add timeout (context?)
// 				RemountFunc(tMounts[i])
// 				return
// 			}
// 			fmt.Printf("ERR - %s\n", err.Error())
// 		}
// 	}
// 	time.Sleep(time.Second * 30)
// 	watch()
// }
