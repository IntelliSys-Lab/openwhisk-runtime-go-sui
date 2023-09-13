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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (ap *ActionProxy) loadHandler(w http.ResponseWriter, r *http.Request) {

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
		// check if you have an action
		if ap.theresnet18Executor == nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet (load)"))
			return
		}
		if ap.theresnet18Executor.started == true {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("already loaded resnet18"))
			return
		}

		NEWresnet18Executor1 := Newresnet18Executor(ap.outFile, ap.errFile, "_test/loadres18.sh", ap.env)
		ap.theresnet18Executor = NEWresnet18Executor1

		//Pre-load libraries & model
		err := ap.theresnet18Executor.Start(false)
		// check for early termination
		if err != nil {
			Debug("WARNING! Command exited (loadHandler)")
			//ap.theresnet18Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("command exited！！"))
			sendError(w, http.StatusBadRequest, fmt.Sprintf(err.Error()))
			return
		}
	} else if strings.Contains(actionName, "ptest05") {
		// check if you have an action
		if ap.theresnet50Executor == nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet (load)"))
			return
		}
		if ap.theresnet50Executor.started == true {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("already loaded resnet50"))
			return
		}

		NEWresnet50Executor1 := Newresnet50Executor(ap.outFile, ap.errFile, "_test/loadres50.sh", ap.env)
		ap.theresnet50Executor = NEWresnet50Executor1

		//Pre-load libraries & model
		err := ap.theresnet50Executor.Start(false)
		// check for early termination
		if err != nil {
			Debug("WARNING! Command exited (load)")
			//ap.theresnet50Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("command exited"))
			return
		}
	} else if strings.Contains(actionName, "ptest06") {
		// check if you have an action
		if ap.theresnet152Executor == nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet (load)"))
			return
		}
		if ap.theresnet152Executor.started == true {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("already loaded resnet152"))
			return
		}

		NEWresnet152Executor1 := Newresnet152Executor(ap.outFile, ap.errFile, "_test/loadres152.sh", ap.env)
		ap.theresnet152Executor = NEWresnet152Executor1

		//Pre-load libraries & model
		err := ap.theresnet152Executor.Start(false)
		// check for early termination
		if err != nil {
			Debug("WARNING! Command exited(load)")
			//ap.theresnet152Executor = nil
			sendError(w, http.StatusBadRequest, fmt.Sprintf("command exited"))
			return
		}
	} else {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Not defined this model!"))
		return
	}
}
