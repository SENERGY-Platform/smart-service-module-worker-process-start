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
	"errors"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"io"
	"net/http"
	"net/url"
)

func (this *ProcessDeploymentStart) Start(token auth.Token, deploymentId string, inputs map[string]interface{}) (instance ProcessInstance, err error) {
	query := ""
	if inputs != nil && len(inputs) > 0 {
		values := url.Values{}
		for key, value := range inputs {
			val, err := json.Marshal(value)
			if err != nil {
				return instance, err
			}
			values.Add(key, string(val))
		}
		query = "?" + values.Encode()
	}
	req, err := http.NewRequest("GET", this.config.ProcessEngineWrapperUrl+"/v2/deployments/"+url.PathEscape(deploymentId)+"/start"+query, nil)
	if err != nil {
		return instance, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return instance, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		return instance, err
	}
	err = json.NewDecoder(resp.Body).Decode(&instance)
	return instance, err
}

type ProcessInstance struct {
	Id             string `json:"id,omitempty"`
	DefinitionId   string `json:"definitionId,omitempty"`
	BusinessKey    string `json:"businessKey,omitempty"`
	CaseInstanceId string `json:"caseInstanceId,omitempty"`
	Ended          bool   `json:"ended,omitempty"`
	Suspended      bool   `json:"suspended,omitempty"`
	TenantId       string `json:"tenantId,omitempty"`
}
