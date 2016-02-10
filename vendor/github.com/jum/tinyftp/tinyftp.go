// copyright 2013 Jens-Uwe Mager jum@anubis.han.de
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tinyftp implements a small subset of FTP, enough to retrieve directory
// listings and files (including resume support). It uses net/textproto to do the heavy
// lifting. Most of the functions return the triple code, message, err besides the real
// result. if err == nil, code and message return the raw informal FTP server code and
// message for the success case.
package tinyftp

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Conn struct {
	conn *textproto.Conn
}

// NewConn returns a new Conn using nconn for I/O.
func NewConn(nconn io.ReadWriteCloser) (conn *Conn, code int, message string, err error) {
	conn = &Conn{
		conn: textproto.NewConn(nconn),
	}
	code, message, err = conn.conn.ReadResponse(2)
	return
}

// Dial connects to the given address on the given network using net.Dial
// and then returns a new Conn for the connection.
func Dial(network, addr string) (conn *Conn, code int, message string, err error) {
	nconn, err := net.Dial(network, addr)
	if err != nil {
		return nil, 0, "", err
	}
	conn, code, message, err = NewConn(nconn)
	return
}

// DialTimeout connects to the given address on the given network with a timeout and
// then returns a new Conn for the connection.
func DialTimeout(network, addr string, timeout time.Duration) (conn *Conn, code int, message string, err error) {
	nconn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, 0, "", err
	}
	conn, code, message, err = NewConn(nconn)
	return
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// Switch underlaying connection to a new instance.
// (Necessary to handle SSL/TLS connections for explicit FTPS servers)
func (c *Conn) SwitchTo(nconn io.ReadWriteCloser) {
	c.conn = textproto.NewConn(nconn)
}

// "AUTH SSL" command to start SSL/TLS on connection.
func (c *Conn) AuthSSL(mode string) (code int, message string, err error) {
	return c.Cmd(234, "AUTH SSL")
}

// Send an FTP command and return the response code and message. The expectCode is given
// as argument to textproto.ReadResponse to signify the expected FTP response code.
func (c *Conn) Cmd(expectCode int, format string, args ...interface{}) (code int, message string, err error) {
	err = c.conn.PrintfLine(format, args...)
	if err != nil {
		return 0, "", err
	}
	code, message, err = c.conn.ReadResponse(expectCode)
	return
}

// Authenticate the session using the given user name and password. Use anonymous@ if they
// are empty.
func (c *Conn) Login(user, pass string) (code int, message string, err error) {
	if len(user) == 0 {
		user = "anonymous"
		if len(pass) == 0 {
			pass = user + "@"
		}
	}
	code, message, err = c.Cmd(0, "USER %s", user)
	if err != nil || code == 230 {
		return code, message, err
	}
	if code != 331 {
		return code, message, &textproto.Error{Code: code, Msg: message}
	}
	code, message, err = c.Cmd(230, "PASS %s", pass)
	return
}

// Set the FTP file transfer type (eg. A or I)
func (c *Conn) Type(mode string) (code int, message string, err error) {
	code, message, err = c.Cmd(200, "TYPE %s", mode)
	if err != nil {
		return code, message, err
	}
	return
}

// Change the working directory.
func (c *Conn) Cwd(dir string) (code int, message string, err error) {
	code, message, err = c.Cmd(250, "CWD %s", dir)
	if err != nil {
		return code, message, err
	}
	return
}

// Terminate the FTP session.
func (c *Conn) Quit() (code int, message string, err error) {
	code, message, err = c.Cmd(221, "QUIT")
	if err != nil {
		return code, message, err
	}
	return
}

var passiveRegexp = regexp.MustCompile(`([\d]+),([\d]+),([\d]+),([\d]+),([\d]+),([\d]+)`)

// Enter passive mode, the addr returned is suitable for net.Dial.
func (c *Conn) Passive() (addr string, code int, message string, err error) {
	code, message, err = c.Cmd(227, "PASV")
	if err != nil {
		return "", code, message, err
	}
	matches := passiveRegexp.FindStringSubmatch(message)
	if matches == nil {
		return "", code, message, fmt.Errorf("Cannot parse PASV response: %s", message)
	}
	ph, err := strconv.Atoi(matches[5])
	if err != nil {
		return "", code, message, err
	}
	pl, err := strconv.Atoi(matches[6])
	if err != nil {
		return "", code, message, err
	}
	port := strconv.Itoa((ph << 8) | pl)
	addr = strings.Join(matches[1:5], ".") + ":" + port
	return
}

// List the specified directory.
func (c *Conn) List(dir string, dconn net.Conn) (dirList []string, code int, message string, err error) {
	defer dconn.Close()

	if len(dir) != 0 {
		code, message, err = c.Cmd(1, "LIST %s", dir)
	} else {
		code, message, err = c.Cmd(1, "LIST")
	}
	if err != nil {
		return nil, code, message, err
	}
	scanner := bufio.NewScanner(dconn)
	for scanner.Scan() {
		dirList = append(dirList, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		return nil, code, message, err
	}
	err = dconn.Close()
	if err != nil {
		return nil, code, message, err
	}

	code, message, err = c.conn.ReadResponse(2)
	return
}

// List the specified directory, names only.
func (c *Conn) NameList(dir string, dconn net.Conn) (dirList []string, code int, message string, err error) {
	defer dconn.Close()

	if len(dir) != 0 {
		code, message, err = c.Cmd(1, "NLST %s", dir)
	} else {
		code, message, err = c.Cmd(1, "NLST")
	}
	if err != nil {
		return nil, code, message, err
	}
	scanner := bufio.NewScanner(dconn)
	for scanner.Scan() {
		dirList = append(dirList, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		return nil, code, message, err
	}
	err = dconn.Close()
	if err != nil {
		return nil, code, message, err
	}

	code, message, err = c.conn.ReadResponse(2)
	return
}

// Return the size of the given file
func (c *Conn) Size(fname string) (size int64, code int, message string, err error) {
	code, message, err = c.Cmd(213, "SIZE %s", fname)
	if err != nil {
		return 0, code, message, err
	}
	size, err = strconv.ParseInt(message, 10, 64)
	return
}

// Start next transfer at the given size
func (c *Conn) Rest(size int64) (code int, message string, err error) {
	code, message, err = c.Cmd(3, "REST %v", size)
	if err != nil {
		return code, message, err
	}
	return
}

// Retrieve the named file
func (c *Conn) Retrieve(fname string, dconn net.Conn) (contents []byte, code int, message string, err error) {
	defer dconn.Close()

	code, message, err = c.Cmd(1, "RETR %s", fname)
	if err != nil {
		return nil, code, message, err
	}
	contents, err = ioutil.ReadAll(dconn)
	if err != nil {
		return nil, code, message, err
	}
	err = dconn.Close()
	if err != nil {
		return nil, code, message, err
	}

	code, message, err = c.conn.ReadResponse(2)
	return
}

// Retrieve the named file to the given io.Writer.
func (c *Conn) RetrieveTo(fname string, dconn net.Conn, w io.Writer) (written int64, code int, message string, err error) {
	defer dconn.Close()

	code, message, err = c.Cmd(1, "RETR %s", fname)
	if err != nil {
		return 0, code, message, err
	}
	written, err = io.Copy(w, dconn)
	if err != nil {
		return 0, code, message, err
	}
	err = dconn.Close()
	if err != nil {
		return 0, code, message, err
	}

	code, message, err = c.conn.ReadResponse(2)
	return
}

// Retrieve the named file to an io.Reader.
func (c *Conn) RetrieveFrom(fname string) (dconn net.Conn, code int, message string, err error) {
	addr, code, message, err := c.Passive()
	if err != nil {
		return nil, code, message, err
	}
	dconn, err = net.Dial("tcp", addr)
	if err != nil {
		return nil, 0, "", err
	}
	code, message, err = c.Cmd(1, "RETR %s", fname)
	if err != nil {
		return nil, code, message, err
	}
	return dconn, code, message, err
}
