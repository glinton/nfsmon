package nfsmon_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/glinton/nfsmon"
)

var localPath = "/tmp/nfsmon-test"

func TestMain(m *testing.M) {
	nfsmon.WatchFreq = time.Second * 1
	nfsmon.RemountFunc = remountFunc

	os.Create(localPath)
	m.Run()
	os.RemoveAll(localPath)
}

func TestWatch(t *testing.T) {
	mount := nfsmon.Mount{
		Server:     "192.168.0.1",
		ServerPath: "/export/thing",
		DestPath:   localPath,
	}

	// test WatchMount
	nfsmon.WatchMount(mount)

	// limit test
	d := time.Now().Add(time.Second * 6)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	// test remount no error
	nfsmon.RemountFunc = remountFunc

	// start watcher
	go nfsmon.Watch(ctx)

	// test remount with error
	time.Sleep(time.Millisecond * 1100)
	nfsmon.ErrCondition = errTrue
	time.Sleep(time.Millisecond * 1100)
	nfsmon.RemountFunc = remountFuncErr

	// test UnwatchMount
	<-ctx.Done()
	nfsmon.UnwatchMount(mount)
}

func remountFunc(m nfsmon.Mount) error {
	cmd := fmt.Sprintf("%s:%s", m.Server, m.ServerPath)
	if m.MountOpts != "" {
		cmd += fmt.Sprintf(",%s", m.MountOpts)
	}
	cmd += fmt.Sprintf(" %s", m.DestPath)
	fmt.Println(cmd)
	return nil
}

func remountFuncErr(m nfsmon.Mount) error {
	return fmt.Errorf("triggerred failure")
}

func errTrue(err error) bool {
	return true
}
