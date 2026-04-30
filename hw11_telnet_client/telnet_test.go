package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("connection refused", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := l.Addr().String()
		require.NoError(t, l.Close())

		client := NewTelnetClient(addr, time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
		err = client.Connect()
		require.Error(t, err)
	})

	t.Run("server closes connection immediately", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer l.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			out := &bytes.Buffer{}
			client := NewTelnetClient(l.Addr().String(), 5*time.Second, io.NopCloser(&bytes.Buffer{}), out)
			require.NoError(t, client.Connect())
			defer client.Close()

			err := client.Receive()
			require.NoError(t, err)
			require.Equal(t, "", out.String())
		}()

		go func() {
			defer wg.Done()
			conn, err := l.Accept()
			require.NoError(t, err)
			conn.Close()
		}()

		wg.Wait()
	})

	t.Run("multiple messages", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer l.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			in := &bytes.Buffer{}
			out := &bytes.Buffer{}
			in.WriteString("line1\nline2\nline3\n")

			client := NewTelnetClient(l.Addr().String(), 5*time.Second, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer client.Close()

			require.NoError(t, client.Send())
			require.NoError(t, client.Receive())
			require.Equal(t, "ok\n", out.String())
		}()

		go func() {
			defer wg.Done()
			conn, err := l.Accept()
			require.NoError(t, err)
			defer conn.Close()

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			require.NoError(t, err)
			require.Equal(t, "line1\nline2\nline3\n", string(buf[:n]))

			_, err = conn.Write([]byte("ok\n"))
			require.NoError(t, err)
		}()

		wg.Wait()
	})

	t.Run("close before connect", func(t *testing.T) {
		client := NewTelnetClient("127.0.0.1:1234", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
		require.NoError(t, client.Close())
	})
}
