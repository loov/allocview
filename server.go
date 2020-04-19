package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"

	"loov.dev/allocview/internal/packet"
	"loov.dev/allocview/internal/series"
)

// ConnectDeadline defines how fast clients should connect to the server.
const ConnectDeadline = 10 * time.Second

// Server is a profile listening server.
type Server struct {
	profiles chan *Profile
}

// NewServer returns a new server.
func NewServer() *Server {
	return &Server{
		profiles: make(chan *Profile, 1024),
	}
}

// Profile returns channel for profiles.
func (server *Server) Profiles() <-chan *Profile { return server.profiles }

// Exec starts listening to cmd.
func (server *Server) Exec(ctx context.Context, group *errgroup.Group, cmd *exec.Cmd) error {
	// create a temporary socket name
	tmpfile, err := ioutil.TempFile("", "alloclog")
	if err != nil {
		return fmt.Errorf("unable to create temporary alloclog file: %w", err)
	}
	sockname := tmpfile.Name()
	tmpfile.Close()
	os.Remove(sockname)

	// setup listener
	addr := &net.UnixAddr{Name: sockname, Net: "unix"}
	sock, err := net.ListenUnix("unix", addr)
	if err != nil {
		return fmt.Errorf("unable to start unix socket on %q: %w", sockname, err)
	}
	sock.SetUnlinkOnClose(true)

	// start the program
	cmd.Env = append(
		os.Environ(),
		"ALLOCLOGSOCK="+sockname,
	)
	err = cmd.Start() // TODO: use pgroup
	if err != nil {
		return fmt.Errorf("failed to start %q: %w", cmd.Args, err)
	}

	// wait for the program to connect
	err = sock.SetDeadline(time.Now().Add(ConnectDeadline))
	if err != nil {
		_ = cmd.Process.Kill()
		_ = sock.Close()
		return fmt.Errorf("failed to set socket deadline: %w", err)
	}

	conn, err := sock.AcceptUnix()
	if err != nil {
		_ = cmd.Process.Kill()
		_ = sock.Close()
		return fmt.Errorf("no connection established, did you import `loov.dev/allocview/attach`: %w", err)
	}

	// we'll set deadline for the first packet to handle misconfigurations
	err = conn.SetReadDeadline(time.Now().Add(ConnectDeadline))
	if err != nil {
		_ = cmd.Process.Kill()
		_ = sock.Close()
		return fmt.Errorf("failed to set read deadline: %w", err)
	}
	var dec packet.Decoder
	err = dec.Read(conn)
	if err != nil {
		return fmt.Errorf("failed to read first packet: %w", err)
	}
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		_ = cmd.Process.Kill()
		_ = sock.Close()
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	// TODO: handle magic header better
	magic := dec.String()
	if magic != "alloclog" {
		return fmt.Errorf("invalid header %q expected %q", magic, "alloclog")
	}

	exename := dec.String()
	funcname := dec.String()
	funcaddr := dec.Uintptr()

	// Reading of profiles.
	group.Go(func() error {
		err := server.readProfiles(conn, exename, funcname, funcaddr)
		log.Printf("readProfiles returned: %v", err)
		return err
	})

	group.Go(func() error {
		// waits for program to close
		err := cmd.Wait()
		log.Printf("program exited: %v", err)
		return err
	})

	return nil
}

func (server *Server) readProfiles(conn *net.UnixConn, exename, funcname string, funcaddr uintptr) error {
	lastState := map[[32]uintptr]series.Sample{}

	var dec packet.Decoder
	for {
		err := dec.Read(conn)
		if err != nil {
			return fmt.Errorf("failed to read packet: %w", err)
		}

		unixnano := dec.Int64()
		count := dec.Uint32()

		profile := &Profile{
			ExeName: exename,

			FuncName: funcname,
			FuncAddr: funcaddr,

			Time: time.Unix(0, unixnano),

			Records: make([]runtime.MemProfileRecord, count),
		}

		for i, rec := range profile.Records {
			var next series.Sample
			next.AllocBytes = dec.Int64()
			next.FreeBytes = dec.Int64()
			next.AllocObjects = dec.Int64()
			next.FreeObjects = dec.Int64()

			for i := 0; ; i++ {
				frame := dec.Uintptr()
				if frame == 0 {
					break
				}

				rec.Stack0[i] = frame
			}

			last, _ := lastState[rec.Stack0]
			lastState[rec.Stack0] = next

			rec.AllocBytes = next.AllocBytes - last.AllocBytes
			rec.FreeBytes = next.FreeBytes - last.FreeBytes
			rec.AllocObjects = next.AllocObjects - last.AllocObjects
			rec.FreeObjects = next.FreeObjects - last.FreeObjects

			profile.Records[i] = rec
		}

		server.profiles <- profile
	}
}

type Profile struct {
	ExeName string

	FuncName string
	FuncAddr uintptr

	Time time.Time

	Records []runtime.MemProfileRecord
}
