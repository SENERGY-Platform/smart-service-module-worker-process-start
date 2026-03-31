/*
 * Copyright (c) 2026 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
)

func (this *ProcessDeploymentStart) CheckFogProcess(token auth.Token, hubId string, businessKey string) (found bool, state State, err error) {
	instances, err := this.GetFogProcessInstances(token, hubId, businessKey)
	if err != nil {
		return found, state, err
	}
	if len(instances) == 0 {
		return false, state, nil
	}
	for _, instance := range instances {
		switch instance.State {
		case "EXTERNALLY_TERMINATED":
			state = Terminated
			found = true
		case "ACTIVE":
			return true, Active, nil
		case "COMPLETED":
			return true, Completed, nil
		}
	}
	return found, state, nil
}

type State int

const (
	Unknown State = iota
	Completed
	Active
	Terminated
)

func (this *ProcessDeploymentStart) CheckProcess(token auth.Token, businessKey string) (found bool, state State, err error) {
	instances, err, _ := client.New(this.config.ProcessEngineWrapperUrl).GetHistoricProcessInstances(token.Jwt(), client.InstanceListOptions{
		BusinessKey: businessKey,
	})
	if err != nil {
		return found, state, err
	}
	if len(instances) == 0 {
		return false, state, nil
	}
	for _, instance := range instances {
		switch instance.State {
		case "EXTERNALLY_TERMINATED":
			state = Terminated
			found = true
		case "ACTIVE":
			return true, Active, nil
		case "COMPLETED":
			return true, Completed, nil
		}
	}
	return found, state, nil
}
