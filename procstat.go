package process_info

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"flag"
	"time"
)

type ProcPidStat struct {
	Pid                   int64
	Comm                  string
	State                 string
	Ppid                  int64
	Pgrp                  int64 // 5
	Session               int64
	Tty_nr                int64
	Tpgid                 int64
	Flags                 uint64
	Minflt                uint64 // 10
	Cminflt               uint64
	Majflt                uint64
	Cmajflt               uint64
	Utime                 uint64
	Stime                 uint64 // 15
	Cutime                int64
	Cstime                int64
	Priority              int64
	Nice                  int64
	Num_threads           int64 // 20
	Itrealvalue           int64
	Starttime             int64
	Vsize                 uint64
	Rss                   int64
	Rsslim                uint64 // 25
	Startcode             uint64
	Endcode               uint64
	Startstack            uint64
	Kstkesp               uint64
	Kstkeip               uint64 // 30
	Signal                uint64
	Blocked               uint64
	Sigignore             uint64
	Sigcatch              uint64
	Wchan                 uint64 // 35
	Nswap                 uint64
	Cnswap                uint64
	Exit_signal           int64
	Processor             int64
	Rt_priority           uint64 // 40
	Policy                uint64
	Delayacct_blkio_ticks uint64
	Guest_time            uint64
	Cguest_tim            int64 // 44
}

var ErrUnexpectedFormat = errors.New("Unexpected format in system file")

func GetProcPidStat(pid int32) (result ProcPidStat, err error) {
	contents, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return
	}

	var left []byte
	var leftleft, right string

	for i := len(contents) - 1; ; i-- {
		if i == 0 {
			err = ErrUnexpectedFormat
			fmt.Println("blah")
			return
		}
		if contents[i] == ')' {
			left = contents[:i]
			right = string(contents[i+1:])
			break
		}
	}

	for i := 0; ; i++ {
		if i >= len(left) {
			err = ErrUnexpectedFormat
			return
		}
		if left[i] == '(' {
			leftleft = string(left[:i])
			result.Comm = string(left[i+1:])
			break
		}
	}

	if n, _ := fmt.Sscan(leftleft, &result.Pid); n != 1 {
		err = ErrUnexpectedFormat
		return
	}
	if n, _ := fmt.Sscan(right,
		&result.State,
		&result.Ppid,
		&result.Pgrp,
		&result.Session,
		&result.Tty_nr,
		&result.Tpgid,
		&result.Flags,
		&result.Minflt,
		&result.Cminflt,
		&result.Majflt,
		&result.Cmajflt,
		&result.Utime,
		&result.Stime,
		&result.Cutime,
		&result.Cstime,
		&result.Priority,
		&result.Nice,
		&result.Num_threads,
		&result.Itrealvalue,
		&result.Starttime,
		&result.Vsize,
		&result.Rss,
		&result.Rsslim,
		&result.Startcode,
		&result.Endcode,
		&result.Startstack,
		&result.Kstkesp,
		&result.Kstkeip,
		&result.Signal,
		&result.Blocked, //30
		&result.Sigignore,
		&result.Sigcatch,
		&result.Wchan,
		&result.Nswap,
		&result.Cnswap,
		&result.Exit_signal,
		&result.Processor,
		&result.Rt_priority,
		&result.Policy,
		&result.Delayacct_blkio_ticks, //40
		&result.Guest_time,
		&result.Cguest_tim); n != 42 { //42

		err = ErrUnexpectedFormat
		return
	}

	err = nil
	return

}

func TotalMemory() (kb uint64, err error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}

	var blah string

	if n, _ := fmt.Fscan(file, &blah, &kb); n != 2 {
		err = ErrUnexpectedFormat
		return
	}
	return
}

func PidMemory(pid int32) (kb uint64, err error) {
	file, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return
	}
	var blah string
	for {
		n, _ := fmt.Fscan(file, &blah)
		if n != 1 {
			err = ErrUnexpectedFormat
			return
		}
		if blah == "VmRSS:" {
			if n, _ := fmt.Fscan(file, &kb); n != 1 {
				err = ErrUnexpectedFormat
				return
			}
			return
		}
	}
	// this isn't really possible.
	err = ErrUnexpectedFormat
	return
}

func TotalCPU() (cpu uint64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return
	}

	var blah string
	var x, y, z uint64
	if n, ee := fmt.Fscan(file, &blah, &x, &y, &z); n != 4 {
		fmt.Println(ee)
		err = ErrUnexpectedFormat
		return
	}

	cpu = x + y + z
	return
}

func PidCPULength(pid int32, d time.Duration) (perc float32, err error) {
	totalcpu, err := TotalCPU()
	if err != nil {
		return
	}
	procstatinfo, err := GetProcPidStat(pid)
	if err != nil {
		return
	}
	time.Sleep(d)
	procstatinfo_after, err := GetProcPidStat(pid)
	if err != nil {
		return
	}
	totalcpu_after, err := TotalCPU()
	if err != nil {
		return
	}
	proc_delta := (procstatinfo_after.Utime + procstatinfo_after.Stime) -
		(procstatinfo.Utime + procstatinfo.Stime)
	totalcpu_delta := totalcpu_after - totalcpu
	if totalcpu_delta <= 0 {
		perc = 0.0
		return
	}
	perc = float32(proc_delta) / float32(totalcpu_delta)
	return
}

func PidCPU(pid int32) (perc float32, err error) {
	return PidCPULength(pid, 100 * time.Millisecond)
}

/*
var pid = flag.Int("pid", 1, "PID to run tests against")
var cpu_len = flag.Duration("sleep", time.Second * 5, "Time to sleep during CPU test")
func main() {
	flag.Parse()
	pid := int32(*pid)
	result, err := GetProcPidStat(pid)
	fmt.Printf("%#v %#v\n", result, err)
	//cpu, err := TotalCPU()
	//fmt.Printf("%d %#v\n", cpu, err)
	fmt.Println(TotalMemory())
	fmt.Println(PidMemory(pid))
	fmt.Println(PidCPULength(pid, *cpu_len))
}
*/