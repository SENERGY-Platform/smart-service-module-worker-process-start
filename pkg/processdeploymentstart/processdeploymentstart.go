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
	"errors"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"log"
	"net/url"
	"runtime/debug"
)

func New(config Config, libConfig configuration.Config, auth *auth.Auth, smartServiceRepo SmartServiceRepo) *ProcessDeploymentStart {
	return &ProcessDeploymentStart{config: config, libConfig: libConfig, auth: auth, smartServiceRepo: smartServiceRepo}
}

type ProcessDeploymentStart struct {
	config           Config
	libConfig        configuration.Config
	auth             *auth.Auth
	smartServiceRepo SmartServiceRepo
}

type SmartServiceRepo interface {
	GetInstanceUser(instanceId string) (userId string, err error)
	UseModuleDeleteInfo(info model.ModuleDeleteInfo) error
	ListExistingModules(processInstanceId string, query model.ModulQuery) (result []model.SmartServiceModule, err error)
}

func (this *ProcessDeploymentStart) Do(task model.CamundaExternalTask) (modules []model.Module, outputs map[string]interface{}, err error) {
	deploymentId := this.getProcessDeploymentId(task)
	if deploymentId == "" {
		return modules, outputs, errors.New("missing process deployment id")
	}
	inputs := this.getProcessStartVariables(task)

	existingModules, err := this.smartServiceRepo.ListExistingModules(task.ProcessInstanceId, model.ModulQuery{})
	if err != nil {
		log.Println("ERROR: unable to get existing modules", err)
		return modules, outputs, err
	}
	userId := ""
	isFog := false
	fogHub := ""
	for _, m := range existingModules {
		userId = m.UserId
		if m.ModuleType == this.config.ProcessDeploymentModuleType && m.ModuleData["process_deployment_id"] == deploymentId {
			var ok bool
			isFog, ok = m.ModuleData["is_fog_deployment"].(bool)
			if !ok {
				isFog = false
			}
			if isFog {
				fogHub, ok = m.ModuleData["fog_hub"].(string)
				if !ok {
					fogHub = ""
				}
			}
		}
	}

	if userId == "" {
		userId, err = this.smartServiceRepo.GetInstanceUser(task.ProcessInstanceId)
		if err != nil {
			log.Println("ERROR: unable to get instance user", err)
			return modules, outputs, err
		}
	}

	token, err := this.auth.ExchangeUserToken(userId)
	if err != nil {
		log.Println("ERROR: unable to exchange user token", err)
		return modules, outputs, err
	}
	if isFog {
		err = this.StartFog(token, fogHub, deploymentId, inputs)
		if err != nil {
			log.Println("ERROR: unable to start fog process", err)
			return modules, outputs, err
		}

		return []model.Module{{
				Id:               this.getModuleId(task),
				ProcesInstanceId: task.ProcessInstanceId,
				SmartServiceModuleInit: model.SmartServiceModuleInit{
					DeleteInfo: nil,
					ModuleType: this.libConfig.CamundaWorkerTopic,
					ModuleData: map[string]interface{}{},
				},
			}},
			map[string]interface{}{},
			err
	} else {
		instance, err := this.Start(token, deploymentId, inputs)
		if err != nil {
			log.Println("ERROR: unable to start process", err)
			return modules, outputs, err
		}
		moduleData := map[string]interface{}{
			"process_instance_id": instance.Id,
		}

		return []model.Module{{
				Id:               this.getModuleId(task),
				ProcesInstanceId: task.ProcessInstanceId,
				SmartServiceModuleInit: model.SmartServiceModuleInit{
					DeleteInfo: &model.ModuleDeleteInfo{
						Url:    this.config.ProcessEngineWrapperUrl + "/v2/process-instances/" + url.PathEscape(instance.Id),
						UserId: userId,
					},
					ModuleType: this.libConfig.CamundaWorkerTopic,
					ModuleData: moduleData,
				},
			}},
			map[string]interface{}{"process_instance_id": instance.Id},
			err
	}
}

func (this *ProcessDeploymentStart) Undo(modules []model.Module, reason error) {
	log.Println("UNDO:", reason)
	for _, module := range modules {
		if module.DeleteInfo != nil {
			err := this.smartServiceRepo.UseModuleDeleteInfo(*module.DeleteInfo)
			if err != nil {
				log.Println("ERROR:", err)
				debug.PrintStack()
			}
		}
	}
}

func (this *ProcessDeploymentStart) getModuleId(task model.CamundaExternalTask) string {
	return task.ProcessInstanceId + "." + task.Id
}
