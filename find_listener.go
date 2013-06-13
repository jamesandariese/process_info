package process_info

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func spltest(c rune) bool {
	switch c {
	case ' ', ':':
		return true
	}
	return false
}

func parseTcpLine(in *bufio.Reader) (port uint16, listening bool, inaddrany bool, socket string, err error) {
	s, err := in.ReadString('\n')
	if err != nil {
		return
	}
	spl := strings.FieldsFunc(s, spltest)
	if spl[5] == "0A" {
		listening = true
	}
	_, err = fmt.Sscanf(spl[2], "%X", &port)
	inaddrany = spl[1] == "00000000" || spl[1] == "00000000000000000000000000000000"
	socket = spl[13]
	return
}

var ErrNoPidWithSocket = errors.New("No PID with the socket found")

func findPidWithSocket(socket string) (pid uint32, err error) {
	targetText := "socket:[" + socket + "]"

	var files []os.FileInfo
	files, err = ioutil.ReadDir("/proc")
	if err != nil {
		return
	}
	for _, file := range files {
		pidstr := file.Name()
		switch pidstr[0] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// it's a PID
		default:
			// it's not a PID, continue the for loop
			continue
		}
		pidFdDir := "/proc/" + pidstr + "/fd/"

		var fds []os.FileInfo
		fds, err = ioutil.ReadDir(pidFdDir)
		if err == nil {
			for _, fd := range fds {
				var linkText string
				linkText, err = os.Readlink(pidFdDir + fd.Name())
				if err == nil {
					if linkText == targetText {
						pidtmp, errtmp := strconv.ParseUint(pidstr, 10, 32)
						pid, err = uint32(pidtmp), errtmp
						return
					}
				}
			}
		}
	}
	err = ErrNoPidWithSocket
	return
}

var ErrNoPidListeningOnPort = errors.New("No process found listening on port")

func findPidListeningOnPortWrap(port uint16, ipv4 bool, inaddranyOnly bool) (pid uint32, err error) {
	var tcpfile *os.File
	if ipv4 {
		tcpfile, err = os.Open("/proc/net/tcp")
		if err != nil {
			return
		}
	} else {
		tcpfile, err = os.Open("/proc/net/tcp6")
		if err != nil {
			return
		}
	}
	defer tcpfile.Close()

	tcp := bufio.NewReader(tcpfile)
	tcp.ReadString('\n')

	for {
		lport, listening, inaddrany, socket, errtmp := parseTcpLine(tcp)
		if errtmp != nil {
			if errtmp != io.EOF {
				err = errtmp
				return
			}
			err = nil
			break
		}
		if (inaddranyOnly && inaddrany ||
			!inaddrany) &&
			port == lport &&
			listening {
			pid, err = findPidWithSocket(socket)
			return // err should be passed through if there is one.
		}
	}
	err = ErrNoPidListeningOnPort
	return
}

func FindPidListeningOnPort(port uint16, inaddranyOnly bool) (pid uint32, err error) {
	pid, err = findPidListeningOnPortWrap(port, true, inaddranyOnly)
	if err == nil {
		if pid > 0 {
			return
		}
	}
	pid, err = findPidListeningOnPortWrap(port, false, inaddranyOnly)
	return
}
/*
func main() {
	//for i := 1; i <= 80; i++ {
	{
		var i uint16 = 22
		pid, err := FindPidListeningOnPort(uint16(i), true)
		if err == nil {
			fmt.Printf("%d is listening on %d\n", pid, i)
		}
	}
}
*/