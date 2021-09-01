package utils

import (
	"io"
	"net"
	"sync"

	"k8s.io/klog"
)

func HandleTCPStream(src io.ReadWriteCloser, remote string) {
	dst, err := net.Dial("tcp", remote)
	if err != nil {
		klog.Errorf("dial remote %s failed, %v", remote, err)
		src.Close()
		return
	}

	s, r := pipe(src, dst)
	klog.V(2).Infof("send remote %s %d, received %d", remote, s, r)
}

func pipe(src io.ReadWriteCloser, dst io.ReadWriteCloser) (int64, int64) {
	var sent, received int64
	var wg sync.WaitGroup
	var o sync.Once
	closeReader := func() {
		_ = src.Close()
		_ = dst.Close()
	}

	wg.Add(2)
	go func() {
		received, _ = io.Copy(src, dst)
		o.Do(closeReader)
		wg.Done()
	}()

	go func() {
		sent, _ = io.Copy(dst, src)
		o.Do(closeReader)
		wg.Done()
	}()

	wg.Wait()
	return sent, received
}
