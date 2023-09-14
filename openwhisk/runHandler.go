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

// ErrResponse is the response when there are errors
type ErrResponse struct {
	Error string `json:"error"`
}

func sendError(w http.ResponseWriter, code int, cause string) {
	errResponse := ErrResponse{Error: cause}
	b, err := json.Marshal(errResponse)
	if err != nil {
		b = []byte("error marshalling error response")
		Debug(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(b)
	w.Write([]byte("\n"))
}

func (ap *ActionProxy) runHandler(w http.ResponseWriter, r *http.Request) {

	// parse the request
	body, err1 := ioutil.ReadAll(r.Body)
	//当使用body, err := ioutil.ReadAll(r.Body)读取r.Body后，会将r.Body的读取位置移动到数据末尾。如果此时在runHandler()函数中再次尝试读取r.Body，将无法获取到任何数据。
	//因此，重置r.Body: Reset r.Body so it can be read again
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	defer r.Body.Close()
	if err1 != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err1))
		return
	}
	Debug("runHandler done reading %d bytes", len(body))

	// remove newlines
	body = bytes.Replace(body, []byte("\n"), []byte(""), -1)
	var response []byte
	var err error

	//proxy本来的设计，是只能给一个action用的。为了支持多个action，我们在执行完inference的action后，刷新executor，重新执行下一次任务。
	//get the action Name
	var req requestBody
	err = json.Unmarshal(body, &req)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err))
		return
	}
	actionName := req.ActionName
	if strings.Contains(actionName, "ptest04") && (ap.theresnet18Executor.started == false) {
		//NEWresnet18Executor1 := Newresnet18Executor(ap.outFile, ap.errFile, "_test/loadres18.sh", ap.env)
		//ap.theresnet18Executor = NEWresnet18Executor1
		Debug("has created res18executor")
		response, err = ap.theOriginresnet18Executor.StartAndWaitForOutput()

		//重建新的executor
		NEWOriginresnet18Executor := NewOriginresnet18Executor(ap.outFile, ap.errFile, "_test/func50.sh", ap.env)
		ap.theOriginresnet18Executor = NEWOriginresnet18Executor

	} else if strings.Contains(actionName, "ptest05") && (ap.theresnet50Executor.started == false) {
		//NEWresnet50Executor1 := Newresnet50Executor(ap.outFile, ap.errFile, "_test/loadres50.sh", ap.env)
		//ap.theresnet50Executor = NEWresnet50Executor1
		Debug("has created res50executor")
		response, err = ap.theOriginresnet50Executor.StartAndWaitForOutput()

		//重建新的executor
		NEWOriginresnet50Executor := NewOriginresnet50Executor(ap.outFile, ap.errFile, "_test/func50.sh", ap.env)
		ap.theOriginresnet50Executor = NEWOriginresnet50Executor

	} else if strings.Contains(actionName, "ptest06") && (ap.theresnet152Executor.started == false) {
		//NEWresnet152Executor1 := Newresnet152Executor(ap.outFile, ap.errFile, "_test/loadres152.sh", ap.env)
		//ap.theresnet152Executor = NEWresnet152Executor1
		Debug("has created res152executor")
		response, err = ap.theOriginresnet152Executor.StartAndWaitForOutput()

		//重建新的executor
		NEWOriginresnet152Executor := NewOriginresnet152Executor(ap.outFile, ap.errFile, "_test/func152.sh", ap.env)
		ap.theOriginresnet152Executor = NEWOriginresnet152Executor

	} else {
		// check if you have an action
		if ap.theExecutor == nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet"))
			return
		}
		// check if the process exited
		if ap.theExecutor.Exited() {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("command exited"))
			return
		}

		response, err = ap.theExecutor.Interact(body)
	}

	// execute the action

	// check for early termination
	if err != nil {
		Debug("WARNING! Command exited (runHandler). Error is: ")
		Debug(string(err.Error()))
		ap.theExecutor = nil
		sendError(w, http.StatusBadRequest, fmt.Sprintf("command exited"))
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

	// diagnostic when you have writing problems
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
		return
	}
	if numBytesWritten != len(response) {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
		return
	}
}
