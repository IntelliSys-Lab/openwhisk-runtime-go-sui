package openwhisk

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// OutputGuard constant string

// Executor is the container and the guardian  of a child process
// It starts a command, feeds input and output, read logs and control its termination
type resnet50Executor struct {
	cmd     *exec.Cmd
	input   io.WriteCloser
	output  *bufio.Reader
	exited  chan bool
	started bool
}

// NewExecutor creates a child subprocess using the provided command line,
// writing the logs in the given file.
// You can then start it getting a communication channel

func Newresnet50Executor(logout *os.File, logerr *os.File, command string, env map[string]string, args ...string) (proc *resnet50Executor) {
	//env:子进程的环境变量; cmd + arg: 真正的命令
	cmd := exec.Command(command, args...) //创建一个可以用来启动命令的 *Cmd
	cmd.Stdout = logout
	cmd.Stderr = logerr
	cmd.Env = []string{} //初始化 *Cmd 的 Env 字段，这个字段用来设置子进程的环境变量
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v) //遍历传入的环境变量 env，并将它们添加到 *Cmd 的 Env 字段
	}
	Debug("env: %v", cmd.Env) //如果 Debugging 是 true，则输出环境变量到调试日志
	if Debugging {
		cmd.Env = append(cmd.Env, "OW_DEBUG=/tmp/action.log")
	}

	input, err := cmd.StdinPipe() //调用 *Cmd 的 StdinPipe 方法，返回一个连接到命令标准输入的管道和一个错误
	if err != nil {
		return nil
	}
	pipeOut, pipeIn, err := os.Pipe()
	if err != nil {
		return nil
	}

	cmd.ExtraFiles = []*os.File{pipeIn} //将 pipeIn 添加到 *Cmd 的 ExtraFiles 字段，这样子进程就可以通过文件描述符访问这个管道
	output := bufio.NewReader(pipeOut)
	//创建一个新的 Executor 并返回。这个 Executor 包括 *Cmd，连接到命令标准输入的管道，
	// 从 pipeOut 读取数据的 *Reader，以及一个 exited 通道，这个通道用来通知命令已经退出。
	return &resnet50Executor{
		cmd,
		input,
		output,
		make(chan bool),
		false,
	}
}

// Start starts the Executor's command and waits for it to be ready to accept input.
// If waitForAck is true, it waits indefinitely for an acknowledgement from the command.
// If waitForAck is false, it waits for a short time to check if the command has exited.
func (proc *resnet50Executor) Start(waitForAck bool) error {
	Debug("Start:")
	err := proc.cmd.Start()
	if err != nil {
		proc.cmd = nil // No need to keep the command around if it failed to start
		return fmt.Errorf("failed to start command: %w", err)
	}

	proc.started = true

	go func() {
		proc.cmd.Wait()
		proc.exited <- true
	}()

	if !waitForAck {
		select {
		case <-proc.exited:
			return fmt.Errorf("command exited!!")
		case <-time.After(3 * time.Second):
			return nil
		}
	}

	// If we reach here, waitForAck is true, so we wait for an acknowledgement from the command
	ack := make(chan error)
	go func() {
		out, err := proc.output.ReadBytes('\n')
		if err != nil {
			ack <- err
			return
		}

		var ackData ActionAck
		err = json.Unmarshal(out, &ackData)
		if err != nil {
			ack <- err
			return
		}

		if !ackData.Ok {
			ack <- fmt.Errorf("The action did not initialize properly.")
			return
		}

		ack <- nil
	}()

	select {
	case err = <-ack:
		return err
	case <-proc.exited:
		return fmt.Errorf("command exited abruptly during initialization")
	}
}

func (proc *resnet50Executor) IsStarted() bool {
	return proc.started
}

// Interact sends input to the Executor's command and waits for output.
// It returns the output from the command, or an error if the command exits before producing output.
func (proc *resnet50Executor) Interact(in []byte) ([]byte, error) {
	//创建一个buffer，即使proc.exited，依然能得到所有output
	//var outputBuffer bytes.Buffer
	//var wg sync.WaitGroup

	_, err := proc.input.Write(in)
	if err != nil {
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}

	chout := make(chan []byte)
	go func() {
		out, err := proc.output.ReadBytes('\n')
		if err != nil {
			chout <- nil
		} else {
			chout <- out
		}
	}()

	//select {
	//case out := <-chout:
	//	if out == nil {
	//		return nil, errors.New("no answer from the action")
	//	}
	//	return out, nil
	//case <-proc.exited:
	//	return nil, errors.New("command exited!!!!!!!!")
	//}
	var finalOut []byte
	for {
		select {
		case out := <-chout:
			if out == nil {
				if len(finalOut) == 0 {
					return nil, errors.New("no answer from the action")
				}
				return finalOut, nil
			}
			finalOut = append(finalOut, out...)
		case <-proc.exited:
			for {
				select {
				case out := <-chout:
					if out != nil {
						finalOut = append(finalOut, out...)
					} else {
						if len(finalOut) == 0 {
							return nil, errors.New("no answer from the action")
						}
						return finalOut, nil
					}
				default:
					if len(finalOut) == 0 {
						return nil, errors.New("command exited!!!!!!!!")
					}
					return finalOut, nil
				}
			}
		}
	}
}

// Stop will kill the process
// and close the channels
func (proc *resnet50Executor) Stop() {
	Debug("stopping")
	if proc.cmd != nil {
		proc.cmd.Process.Kill()
		proc.cmd = nil
	}
}
