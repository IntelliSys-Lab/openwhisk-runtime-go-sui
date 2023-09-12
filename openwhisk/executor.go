/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
const OutputGuard = "XXX_THE_END_OF_A_WHISK_ACTIVATION_XXX\n"

// DefaultTimeoutStart to wait for a process to start
var DefaultTimeoutStart = 5 * time.Millisecond

// Executor is the container and the guardian  of a child process
// It starts a command, feeds input and output, read logs and control its termination
type Executor struct {
	cmd    *exec.Cmd
	input  io.WriteCloser
	output *bufio.Reader
	exited chan bool
}

// NewExecutor creates a child subprocess using the provided command line,
// writing the logs in the given file.
// You can then start it getting a communication channel

func NewExecutor(logout *os.File, logerr *os.File, command string, env map[string]string, args ...string) (proc *Executor) {
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
	return &Executor{ //创建一个新的 Executor 并返回。这个 Executor 包括 *Cmd，连接到命令标准输入的管道，
		// 从 pipeOut 读取数据的 *Reader，以及一个 exited 通道，这个通道用来通知命令已经退出。
		cmd,
		input,
		output,
		make(chan bool),
	}
}

// Interact interacts with the underlying process
func (proc *Executor) Interact(in []byte) ([]byte, error) {
	// input to the subprocess
	proc.input.Write(in)           //将输入的字节切片 in 写入到子进程的输入流中
	proc.input.Write([]byte("\n")) //向子进程的输入流中写入一个换行符，表示输入结束

	chout := make(chan []byte) //创建一个用于接收子进程输出的通道

	go func() {
		out, err := proc.output.ReadBytes('\n') //启动一个并发的 goroutine，它尝试从子进程的输出流中读取数据，直到遇到换行符
		if err == nil {
			chout <- out //如果读取成功，将读取的数据发送到 chout 通道
		} else {
			chout <- []byte{} //否则,发送一个空的字节切片
		}
	}()
	var err error
	var out []byte
	select { //等待从 chout 通道接收数据或者 proc.exited 通道接收数据，表示子进程已经退出
	case out = <-chout:
		if len(out) == 0 { //如果从 chout 通道接收到的数据长度为 0，表示子进程没有返回任何数据
			err = errors.New("no answer from the action")
		}
	case <-proc.exited: //如果子进程已经退出
		err = errors.New("command exited")
	}
	proc.cmd.Stdout.Write([]byte(OutputGuard)) //在子进程的标准输出和标准错误流中写入结束标记。
	proc.cmd.Stderr.Write([]byte(OutputGuard))
	return out, err
}

// Exited checks if the underlying command exited
func (proc *Executor) Exited() bool {
	select {
	case <-proc.exited:
		return true
	default:
		return false
	}
}

// ActionAck is the expected data structure for the action acknowledgement
type ActionAck struct {
	Ok bool `json:"ok"`
}

// Start execution of the command
// if the flag ack is true, wait forever for an acknowledgement
// if the flag ack is false wait a bit to check if the command exited
// returns an error if the program fails
func (proc *Executor) Start(waitForAck bool) error {
	// start the underlying executable
	Debug("Start:")
	err := proc.cmd.Start() //调用 *Cmd 的 Start 方法，开始执行命令
	if err != nil {         //如果有错误，输出错误信息并返回一个 "command exited" 的错误
		Debug("run: early exit")
		proc.cmd = nil // no need to kill
		return fmt.Errorf("command exited")
	}
	Debug("pid: %d", proc.cmd.Process.Pid) //如果 Debugging 是 true，则在调试日志中输出命令的进程 ID

	go func() { //启动一个并发的 goroutine，它等待命令的结束，然后向 proc.exited 通道发送 true
		proc.cmd.Wait()
		proc.exited <- true
	}()

	// not waiting for an ack, so use a timeout
	if !waitForAck { //如果 waitForAck 是 false，则等待命令的结束或者一个默认的超时时间
		select {
		case <-proc.exited:
			return fmt.Errorf("command exited!!!!")
		case <-time.After(DefaultTimeoutStart):
			return nil
		}
	}

	// wait for acknowledgement
	Debug("waiting for an ack")
	ack := make(chan error)
	go func() { //启动一个并发的 goroutine，它等待子进程退出，然后向 proc.exited 通道发送 true
		out, err := proc.output.ReadBytes('\n')
		Debug("received ack %s", out)
		if err != nil {
			ack <- err
			return
		}
		// parse ack
		var ackData ActionAck
		err = json.Unmarshal(out, &ackData)
		if err != nil {
			ack <- err
			return
		}
		// check ack
		if !ackData.Ok {
			ack <- fmt.Errorf("The action did not initialize properly.")
			return
		}
		ack <- nil
	}()
	// wait for ack or unexpected termination
	select {
	// ack received
	case err = <-ack:
		return err
	// process exited
	case <-proc.exited:
		return fmt.Errorf("Command exited abruptly during initialization.")
	}
}

// Stop will kill the process
// and close the channels
func (proc *Executor) Stop() {
	Debug("stopping")
	if proc.cmd != nil {
		proc.cmd.Process.Kill()
		proc.cmd = nil
	}
}
