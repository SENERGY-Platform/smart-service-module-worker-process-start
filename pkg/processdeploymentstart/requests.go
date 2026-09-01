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
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/gin-middleware/otelx"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
)

func (this *ProcessDeploymentStart) Start(ctx context.Context, token auth.Token, deploymentId string, inputs map[string]interface{}, businessKey string) (instance ProcessInstance, err error) {
	values := url.Values{}
	values.Add("business_key", businessKey)
	for key, value := range inputs {
		val, err := json.Marshal(value)
		if err != nil {
			return instance, err
		}
		values.Add(key, string(val))
	}
	query := "?" + values.Encode()
	req, err := http.NewRequest("GET", this.config.ProcessEngineWrapperUrl+"/v2/deployments/"+url.PathEscape(deploymentId)+"/start"+query, nil)
	if err != nil {
		return instance, err
	}
	err = otelx.InjectContextToRequest(ctx, req)
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

func (this *ProcessDeploymentStart) StartFog(ctx context.Context, token auth.Token, hubId string, deploymentId string, inputs map[string]interface{}, businessKey string) error {
	values := url.Values{}
	values.Add("business_key", businessKey)
	for key, value := range inputs {
		val, err := json.Marshal(value)
		if err != nil {
			return err
		}
		values.Add(key, string(val))
	}
	query := "?" + values.Encode()
	req, err := http.NewRequest("GET", this.config.FogProcessDeploymentUrl+"/deployments/"+url.PathEscape(hubId)+"/"+url.PathEscape(deploymentId)+"/start"+query, nil)
	if err != nil {
		return err
	}
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		return err
	}
	return err
}

func (this *ProcessDeploymentStart) GetFogProcessInstances(ctx context.Context, token auth.Token, hubId string, businessKey string) (result model.HistoricProcessInstances, err error) {
	query := url.Values{"network_id": {hubId}, "business_key": {businessKey}}
	req, err := http.NewRequest("GET", this.config.ProcessSyncUrl+"/history/process-instances?"+query.Encode(), nil)
	if err != nil {
		return result, err
	}
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(resp.Status + ": " + string(temp))
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
