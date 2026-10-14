package jsruntime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// HTTPStream owns an HTTP response until EOF, cancellation, or its deadline.
// Reads are pulled by the consumer; no body chunks accumulate in the VM queue.
type HTTPStream struct {
	ctx      context.Context
	cancel   context.CancelFunc
	ready    chan struct{}
	mu       sync.Mutex
	readMu   sync.Mutex
	response *http.Response
	err      error
	bytes    int64
}

func NewHTTPStream(parent context.Context, url, method, headersJSON, bodyHex string, allowInsecure bool) *HTTPStream {
	ctx, cancel := context.WithCancel(parent)
	s := &HTTPStream{ctx: ctx, cancel: cancel, ready: make(chan struct{})}
	go func() {
		response, requestCancel, err := startHTTPRequest(ctx, url, method, headersJSON, bodyHex, allowInsecure)
		s.mu.Lock()
		if ctx.Err() != nil && response != nil {
			response.Body.Close()
			requestCancel()
			response = nil
			err = ctx.Err()
		}
		s.response, s.err = response, err
		if requestCancel != nil {
			// The child request deadline covers the body as well as the headers.
			context.AfterFunc(ctx, requestCancel)
		}
		s.mu.Unlock()
		close(s.ready)
	}()
	return s
}

func (s *HTTPStream) Headers() (int, http.Header, error) {
	select {
	case <-s.ctx.Done():
		return 0, nil, s.ctx.Err()
	case <-s.ready:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return 0, nil, s.err
	}
	if s.ctx.Err() != nil {
		return 0, nil, s.ctx.Err()
	}
	return s.response.StatusCode, s.response.Header.Clone(), nil
}

// Read returns at most 16 KiB; nil means EOF. Concurrent reads are rejected.
func (s *HTTPStream) Read() ([]byte, error) {
	if !s.readMu.TryLock() {
		return nil, fmt.Errorf("HTTP stream read already pending")
	}
	defer s.readMu.Unlock()
	if _, _, err := s.Headers(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	body := s.response.Body
	s.mu.Unlock()
	buf := make([]byte, 16<<10)
	n, err := body.Read(buf)
	s.bytes += int64(n)
	if s.bytes > maxFetchBodyBytes {
		s.Close()
		return nil, fmt.Errorf("HTTP stream exceeds %d MiB", maxFetchBodyBytes>>20)
	}
	if err != nil && err != io.EOF {
		s.Close()
		return nil, err
	}
	// Return a final nonempty chunk before EOF; the next read completes the stream.
	if n > 0 {
		return buf[:n], nil
	}
	if err == io.EOF {
		s.Close()
		return nil, nil
	}
	return []byte{}, nil
}

func (s *HTTPStream) Close() {
	s.cancel()
	<-s.ready
	s.mu.Lock()
	if s.response != nil {
		s.response.Body.Close()
	}
	s.mu.Unlock()
}
