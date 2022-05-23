//go:build !windows
// +build !windows

/**
 * @Description
 * @Author r0cky
 * @Date 2021/10/7 23:14
 **/
package job

import (
	"bytes"
	"context"
	"github.com/panjf2000/ants/v2"
	"github.com/r0ckysec/go-security/bin/misc"
	system "github.com/r0ckysec/go-security/job/util"
	"github.com/r0ckysec/go-security/secio"
	"github.com/thinkeridea/go-extend/exbytes"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

//Job 任务
type Job struct {
	ID      string
	Status  int
	Cmd     string
	Dir     string
	Msg     string
	ExecCmd *exec.Cmd
	cancel  context.CancelFunc
}

// Start 可以为空，可以传入多个参数，仅限传入交互命令和超时配置
func (j *Job) Start(config ...map[string]interface{}) (err error) {
	Jobmap.Set(j.ID, j)
	defer Jobmap.Remove(j.ID)
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout, ok := config[0]["timeout"].(int); ok {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	j.cancel = cancel
	//var stdoutBuf, stderrBuf bytes.Buffer
	stdoutBuf := secio.Buffer.Get().(*bytes.Buffer)
	stdoutBuf.Reset()
	defer func() {
		if stdoutBuf != nil {
			secio.Buffer.Put(stdoutBuf)
			stdoutBuf = nil
		}
	}()
	stderrBuf := secio.Buffer.Get().(*bytes.Buffer)
	stderrBuf.Reset()
	defer func() {
		if stderrBuf != nil {
			secio.Buffer.Put(stderrBuf)
			stderrBuf = nil
		}
	}()
	var cmd *exec.Cmd
	sysType := runtime.GOOS
	if sysType == "linux" {
		cmd = exec.CommandContext(ctx, "bash", "-c", j.Cmd)
	} else if sysType == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", j.Cmd)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", j.Cmd)
	}
	if j.Dir != "" {
		cmd.Dir = j.Dir
	}
	system.SetPgid(cmd)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	//进入交互后的命令
	if len(config) > 0 {
		if in, ok := config[0]["in"].(string); ok {
			inPipe, _ := cmd.StdinPipe()
			_, _ = inPipe.Write(misc.Str2Bytes(in))
			_, _ = inPipe.Write([]byte("\n"))
			_ = inPipe.Close()
		}
	}
	j.ExecCmd = cmd
	//stdoutIn, _ := cmd.StdoutPipe()
	//stderrIn, _ := cmd.StderrPipe()
	//err = cmd.Start()
	//if err != nil {
	//	j.Status = -1
	//	j.Msg = err.Error()
	//	//log.Log.Println("Start Err: ", err.Error())
	//	return
	//}
	//if cmd.Process != nil {
	//	//var finish = make(chan struct{})
	//	//defer close(finish)
	//	//go func() {
	//	//	select {
	//	//	case <-ctx.Done(): //超时/被 cancel 结束
	//	//		//kill -(-PGID) 杀死整个进程组
	//	//		//把任务停掉
	//	//		j.SendKill()
	//	//		if j.ExecCmd.Process != nil {
	//	//			j.KillPid(j.ExecCmd.Process.Pid)
	//	//		}
	//	//		// 防止产生僵尸进程
	//	//		_ = j.ExecCmd.Wait()
	//	//	case <-finish: //正常结束
	//	//	}
	//	//}()
	//	//log.Log.Println("Start child process with pid", cmd.Process.Pid)
	//}
	// 防止产生僵尸进程
	if err := cmd.Run(); err != nil {
		if j.Status == 3 {
			j.Msg = ""
			return nil
		}
		j.Status = -1
		if stderrBuf.Len() > 0 {
			j.Msg = exbytes.ToString(stderrBuf.Bytes())
		} else {
			j.Msg = err.Error()
		}
		//log.Log.Printf("Child process %d exit with err: %v\n", cmd.Process.Pid, err)
	} else {
		if j.Status == 3 {
			j.Msg = ""
			return nil
		}
		if stderrBuf.Len() > 0 {
			j.Status = -1
			j.Msg = exbytes.ToString(stderrBuf.Bytes())
		} else {
			j.Status = 2
			j.Msg = ""
		}
		if strings.Contains(strings.ToLower(j.ID), "check") {
			stdoutBuf.Write(stderrBuf.Bytes())
			j.Msg = exbytes.ToString(stdoutBuf.Bytes())
		}
	}
	return nil
}

func (j *Job) Stop() {
	j.Status = 3
	defer func() {
		j.cancel()
		Jobmap.Remove(j.ID)
	}()
	//把任务停掉
	if j.ExecCmd != nil {
		if j.ExecCmd.Process != nil {
			j.KillPid(j.ExecCmd.Process.Pid)
			_ = j.ExecCmd.Process.Release()
			_, _ = j.ExecCmd.Process.Wait()
		}
		// 防止产生僵尸进程
		_ = j.ExecCmd.Wait()
		//_, _ = j.ExecCmd.Process.Wait()
	}
}

func (j *Job) KillPid(pid int) {
	//log.Log.Printf("kill 进程 pid: %d\n", pid)
	err := system.KillAll(pid)
	if err != nil {
		//log.Log.Printf("kill 进程失败. pid: %d\n", pid)
	}
}

func (j *Job) SendKill() {
	j.Status = 3
	c := j.ExecCmd
	if c.Process != nil {
		//p, _ := os.FindProcess(c.Process.Pid)
		//_ = p.Signal(syscall.SIGINT)
		err := c.Process.Signal(syscall.SIGINT)
		if err != nil {
			//log.Log.Printf("通知 pid: %d 子线程主动结束失败\n", p.Pid)
		} else {
			//log.Log.Printf("已通知 pid: %d 子线程主动结束\n", p.Pid)
		}
		//防止产生僵尸进程
		//_ = c.Wait()
		_, _ = c.Process.Wait()
		j.Stop()
		//if state, _ := c.Process.Wait(); state != nil {
		//	//log.Log.Printf("Child process %d exit with state: %s\n", p.Pid, state.String())
		//}
	}
}

func StopJobAll() {
	//把所有任务停掉
	//log.Log.Println("stop jobs")
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(Jobmap.Count(), func(i interface{}) {
		defer wg.Done()
		i.(*Job).SendKill()
	})
	defer p.Release()
	for item := range Jobmap.IterBuffered() {
		wg.Add(1)
		_ = p.Invoke(item.Val)
		//job := item.Val.(*Job)
		//job.Stop()
	}
	wg.Wait()
}

func GetJob(id string) *Job {
	if job, ok := Jobmap.Get(id); ok {
		j, _ := job.(*Job)
		return j
	} else {
		//log.Log.Printf("%s任务不存在。\n", id)
		return nil
	}
}
