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

func (ap *ActionProxy) loadRunHandler(w http.ResponseWriter, r *http.Request) {

	//if path == "run"，也执行loadHandler
	//在最开始，先得到r.Body.ActionName，然后分析：是否有对应的modelExecutor（通过IsStarted参数）。
	//if IsStarted == false：证明是冷启动，直接调用runHandler(w,r）
	//else：证明已经有pre-load的容器了，继续往下执行。

	// parse the request
	body, err := ioutil.ReadAll(r.Body)
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
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet"))
			return
		}
		if ap.theresnet18Executor.started == false { //没有预加载，所以直接进入cold start
			ap.runHandler(w, r)
			return
		}

		//停止其他model的进程
		//ap.theresnet50Executor.Stop()
		//ap.theresnet152Executor.Stop()

		// remove newlines
		body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

		// execute the action
		response, err := ap.theresnet18Executor.Interact(body)

		// check for early termination
		if err != nil {
			Debug("WARNING! Command exited")
			ap.theresnet18Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("command exited"))
			return
		}
		DebugLimit("received:", response, 120)

		// check if the answer is an object map
		var objmap map[string]*json.RawMessage
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
	} else if strings.Contains(actionName, "ptest05") {
		// actionName contains "ptest"
		Debug("done reading %d bytes", len(body))

		// check if you have an action
		if ap.theresnet50Executor == nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet"))
			return
		}
		if ap.theresnet50Executor.started == false {
			ap.runHandler(w, r)
			return
		}

		//停止其他model的进程
		//ap.theresnet18Executor.Stop()
		//ap.theresnet152Executor.Stop()

		// remove newlines
		body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

		// execute the action
		response, err := ap.theresnet50Executor.Interact(body)

		// check for early termination
		if err != nil {
			Debug("WARNING! Command exited")
			ap.theresnet50Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("resnet18 command exited"))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))
			return
		}
		DebugLimit("received:", response, 120)

		// check if the answer is an object map
		var jsonObj map[string]interface{}
		err = json.Unmarshal(response, &jsonObj)
		//
		//var objmap map[string]*json.RawMessage
		//err = json.Unmarshal(response, &objmap)
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
	} else if strings.Contains(actionName, "ptest06") {
		// actionName contains "ptest"
		Debug("done reading %d bytes", len(body))

		// check if you have an action
		if ap.theresnet152Executor == nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet"))
			return
		}
		if ap.theresnet152Executor.started == false {
			ap.runHandler(w, r)
			return
		}

		//停止其他model的进程
		//ap.theresnet18Executor.Stop()
		//ap.theresnet50Executor.Stop()

		// remove newlines
		body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

		// execute the action
		response, err := ap.theresnet152Executor.Interact(body)

		sendError(w, http.StatusBadGateway, "Response is: ")
		sendError(w, http.StatusBadGateway, string(response))
		sendError(w, http.StatusBadGateway, "ERROR is: ")
		sendError(w, http.StatusBadGateway, err.Error())

		// check for early termination
		if err != nil {
			Debug("WARNING! Command exited")
			ap.theresnet152Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("command exited"))
			return
		}
		DebugLimit("received:", response, 120)

		// check if the answer is an object map
		var objmap map[string]*json.RawMessage
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
	} else {
		ap.runHandler(w, r)
	}
}
