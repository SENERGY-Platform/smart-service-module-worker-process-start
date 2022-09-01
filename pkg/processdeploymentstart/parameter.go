/*
 * Copyright (c) 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package processdeploymentstart

import (
	"encoding/json"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"strings"
)

func (this *ProcessDeploymentStart) getProcessDeploymentId(task model.CamundaExternalTask) string {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"process_deployment_id"]
	if !ok {
		return ""
	}
	result, ok := variable.Value.(string)
	if !ok {
		return ""
	}
	return result
}

func (this *ProcessDeploymentStart) getProcessStartVariables(task model.CamundaExternalTask) (result map[string]interface{}) {
	prefix := this.config.WorkerParamPrefix + "input."
	for keyWithPrefix, value := range task.Variables {
		if strings.HasPrefix(keyWithPrefix, prefix) {
			key := strings.TrimPrefix(keyWithPrefix, prefix)
			str, ok := value.Value.(string)
			if ok {
				if result == nil {
					result = map[string]interface{}{}
				}
				var temp interface{}
				err := json.Unmarshal([]byte(str), &temp)
				if err != nil {
					result[key] = str
				} else {
					result[key] = temp
				}
			}
		}
	}
	return result
}
