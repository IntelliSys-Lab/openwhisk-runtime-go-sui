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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type requestBody struct {
	ActionName string `json:"action_name"`
	// Include any other fields you need
}

type Data struct {
	Content string `json:"content"`
}

func (ap *ActionProxy) loadRunHandler(w http.ResponseWriter, r *http.Request) {

	//在最开始，先得到r.Body.ActionName，然后分析：是否有对应的modelExecutor（通过IsStarted参数）。
	//if IsStarted == false：证明是冷启动，直接调用runHandler(w,r）
	//else：证明已经有pre-load的容器了，继续往下执行。

	// parse the request
	body, err := ioutil.ReadAll(r.Body)

	//当使用body, err := ioutil.ReadAll(r.Body)读取r.Body后，会将r.Body的读取位置移动到数据末尾。如果此时在runHandler()函数中再次尝试读取r.Body，将无法获取到任何数据。
	//因此，重置r.Body: Reset r.Body so it can be read again
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	defer r.Body.Close()
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	//get the action Name
	var req requestBody
	err = json.Unmarshal(body, &req)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err))
		return
	}
	actionName := req.ActionName

	if strings.Contains(actionName, "ptest04") {
		// actionName contains "ptest"
		Debug("done reading %d bytes", len(body))

		// check if you have an action
		if ap.theresnet18Executor == nil {
			//sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet (new)"))
			NEWresnet18Executor1 := Newresnet18Executor(ap.outFile, ap.errFile, "_test/loadres18.sh", ap.env)
			ap.theresnet18Executor = NEWresnet18Executor1
			ap.runHandler(w, r)
			return
		}
		if ap.theresnet18Executor.started == false {
			ap.runHandler(w, r)
			return
		}

		////停止其他model的进程
		//if ap.theresnet50Executor.started == true {
		//	ap.theresnet50Executor.Stop()
		//}
		//if ap.theresnet152Executor.started == true {
		//	ap.theresnet152Executor.Stop()
		//}

		// remove newlines
		body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

		// execute the action
		response, err := ap.theresnet18Executor.Interact(body)

		// check for early termination
		if err != nil {
			Debug("WARNING! RES18 Command exited！！")
			ap.theresnet18Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("resnet18 command exited"))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))
			return
		}
		DebugLimit("received:", response, 120)

		// check if the answer is an object map
		var objmap map[string]*json.RawMessage
		resStr := strings.ReplaceAll(string(response), "'", "\"")
		response = []byte(resStr)
		err = json.Unmarshal(response, &objmap)
		if err != nil {
			sendError(w, http.StatusBadGateway, "The action did not return a dictionary.")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response)))
		numBytesWritten, err := w.Write(response)

		// flush output
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		if ap.theresnet18Executor == nil {
			NEWresnet18Executor1 := Newresnet18Executor(ap.outFile, ap.errFile, "_test/loadres18.sh", ap.env)
			ap.theresnet18Executor = NEWresnet18Executor1
		}

		if ap.theresnet18Executor.started == true {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("already loaded resnet18"))

		}
		err2 := ap.theresnet18Executor.Start(false)
		if err2 != nil {
			Debug("WARNING! Command exited")
			//ap.theresnet18Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Res18 command exited！！"))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))

		}

		// diagnostic when you have writing problems
		if err != nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
			return
		}
		if numBytesWritten != len(response) {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
			return
		}
	} else if strings.Contains(actionName, "ptest05") {
		// actionName contains "ptest"
		Debug("done reading %d bytes", len(body))

		// check if you have an action
		if ap.theresnet50Executor == nil {
			//sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet (new)"))
			NEWresnet50Executor1 := Newresnet50Executor(ap.outFile, ap.errFile, "_test/loadres50.sh", ap.env)
			ap.theresnet50Executor = NEWresnet50Executor1
			ap.runHandler(w, r)
			return
		}
		if ap.theresnet50Executor.started == false {
			ap.runHandler(w, r)
		}

		////停止其他model的进程
		//if ap.theresnet18Executor.started == true {
		//	ap.theresnet18Executor.Stop()
		//}
		//if ap.theresnet152Executor.started == true {
		//	ap.theresnet152Executor.Stop()
		//}

		// remove newlines
		body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

		// execute the action
		response, err := ap.theresnet50Executor.Interact(body)

		// check for early termination
		if err != nil {
			Debug("WARNING!!! Command exited")
			//ap.theresnet50Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("resnet50 command exited"))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(string(response)))
			return
		}
		DebugLimit("received:", response, 120)

		// check if the answer is an object map
		var objmap map[string]*json.RawMessage
		resStr := strings.ReplaceAll(string(response), "'", "\"")
		response = []byte(resStr)
		err = json.Unmarshal(response, &objmap)
		if err != nil {
			sendError(w, http.StatusBadGateway, "The action did not return a dictionary.")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response)))
		numBytesWritten, err := w.Write(response)

		// flush output
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		if ap.theresnet50Executor == nil {
			Debug("yes, executor is nil")
			NEWresnet50Executor1 := Newresnet50Executor(ap.outFile, ap.errFile, "_test/loadres50.sh", ap.env)
			ap.theresnet50Executor = NEWresnet50Executor1
		}

		if ap.theresnet50Executor.started == true {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("already loaded resnet50"))
		}

		err2 := ap.theresnet50Executor.Start(false)
		if err2 != nil {
			Debug("WARNING! Command exited?")
			//ap.theresnet18Executor = nil
			//sendError(w, http.StatusBadRequest, fmt.Sprintf("Res50 command exited！！"))
			//sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))
			Debug(fmt.Sprintf(err.Error()))
		}

		// diagnostic when you have writing problems
		if err != nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
			return
		}
		if numBytesWritten != len(response) {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
			return
		}
	} else if strings.Contains(actionName, "ptest06") {
		// actionName contains "ptest"
		Debug("done reading %d bytes", len(body))

		// check if you have an action
		if ap.theresnet152Executor == nil {
			//sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet (new)"))
			NEWresnet152Executor1 := Newresnet152Executor(ap.outFile, ap.errFile, "_test/loadres152.sh", ap.env)
			ap.theresnet152Executor = NEWresnet152Executor1
			ap.runHandler(w, r)
			return
		}
		if ap.theresnet152Executor.started == false {
			ap.runHandler(w, r)
			return
		}

		////停止其他model的进程
		//if ap.theresnet18Executor.started == true {
		//	ap.theresnet18Executor.Stop()
		//}
		//if ap.theresnet50Executor.started == true {
		//	ap.theresnet50Executor.Stop()
		//}

		// remove newlines
		body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

		// execute the action
		response, err := ap.theresnet152Executor.Interact(body)

		// check for early termination
		if err != nil {
			Debug("WARNING! Command exited")
			ap.theresnet152Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("resnet152 command exited"))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))
			return
		}
		DebugLimit("received:", response, 120)

		// check if the answer is an object map
		var objmap map[string]*json.RawMessage
		resStr := strings.ReplaceAll(string(response), "'", "\"")
		response = []byte(resStr)
		err = json.Unmarshal(response, &objmap)
		if err != nil {
			sendError(w, http.StatusBadGateway, "The action did not return a dictionary.")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response)))
		numBytesWritten, err := w.Write(response)

		// flush output
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		if ap.theresnet152Executor == nil {
			//sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet (new)"))
			NEWresnet152Executor1 := Newresnet152Executor(ap.outFile, ap.errFile, "_test/loadres152.sh", ap.env)
			ap.theresnet152Executor = NEWresnet152Executor1
		}

		if ap.theresnet152Executor.started == true {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("already loaded resnet152"))

		}
		err2 := ap.theresnet152Executor.Start(false)
		if err2 != nil {
			Debug("WARNING! Command exited")
			//ap.theresnet18Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Res152 command exited！！"))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))
		}

		// diagnostic when you have writing problems
		if err != nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
			return
		}
		if numBytesWritten != len(response) {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
			return
		}
	} else {
		ap.runHandler(w, r)
	}
}
