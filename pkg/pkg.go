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

package pkg

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	lib "github.com/SENERGY-Platform/smart-service-module-worker-lib"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/camunda"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/smartservicerepository"
	"github.com/SENERGY-Platform/smart-service-module-worker-process-start/pkg/processdeploymentstart"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config processdeploymentstart.Config, libConfig configuration.Config) error {
	handlerFactory := func(auth *auth.Auth, smartServiceRepo *smartservicerepository.SmartServiceRepository) (camunda.Handler, error) {

		interval, err := time.ParseDuration(config.HealthCheckInterval)
		if err != nil {
			return nil, err
		}

		handler := processdeploymentstart.New(config, libConfig, auth, smartServiceRepo)

		healthCheck := func(module model.SmartServiceModule) (health error, err error) {
			token, err := auth.ExchangeUserToken(module.UserId)
			if err != nil {
				return nil, err
			}
			isFogDeployment, fogHubId, businessKey, state, err := getInstanceBusinessKey(module.ModuleData)
			if errors.Is(err, MissingBusinessKeyError) {
				return nil, nil //old deployments can't be checked
			}
			if err != nil {
				return nil, err
			}
			if state == processdeploymentstart.Completed {
				return nil, nil //completed instances are not checked
			}
			var found bool
			if isFogDeployment {
				found, state, err = handler.CheckFogProcess(token, fogHubId, businessKey)
			} else {
				found, state, err = handler.CheckProcess(token, businessKey)
			}

			if err != nil {
				return nil, err
			}
			if !found {
				return fmt.Errorf("process instance with business key %v not found", businessKey), nil
			}

			if state == processdeploymentstart.Active {
				return nil, nil
			}
			if state == processdeploymentstart.Terminated {
				return fmt.Errorf("process instance with business key %v is terminated", businessKey), nil
			}
			if state == processdeploymentstart.Completed {
				moduleUpdate := model.Module{
					Id:                     module.Id,
					ProcesInstanceId:       module.InstanceId,
					SmartServiceModuleInit: module.SmartServiceModuleInit,
				}
				moduleUpdate.ModuleData["state"] = "completed"
				_, err = smartServiceRepo.SendWorkerModule(moduleUpdate)
				if err != nil {
					return nil, err
				}
				return nil, nil
			}

			return nil, nil
		}
		moduleQuery := model.ModulQuery{TypeFilter: &libConfig.CamundaWorkerTopic}
		smartServiceRepo.StartHealthCheck(ctx, interval, moduleQuery, healthCheck) //timer loop
		smartServiceRepo.RunHealthCheck(moduleQuery, healthCheck)                  //initial check

		return handler, nil
	}
	return lib.Start(ctx, wg, libConfig, handlerFactory)
}

var MissingBusinessKeyError = fmt.Errorf("missing business_key in module data")

func getInstanceBusinessKey(moduleData map[string]interface{}) (isFogDeployment bool, fogHubId string, businessKey string, state processdeploymentstart.State, err error) {
	stateStr, ok := moduleData["state"].(string)
	if ok {
		switch stateStr {
		case "active":
			state = processdeploymentstart.Active
		case "terminated":
			state = processdeploymentstart.Terminated
		case "completed":
			state = processdeploymentstart.Completed
		default:
			state = processdeploymentstart.Unknown
		}
	} else {
		state = processdeploymentstart.Unknown
	}
	businessKey, ok = moduleData["business_key"].(string)
	if !ok {
		return false, "", businessKey, state, MissingBusinessKeyError
	}
	fogHubIdInterface, ok := moduleData["fog_hub"]
	if !ok {
		return false, "", businessKey, state, nil
	}
	fogHubId, ok = fogHubIdInterface.(string)
	if !ok {
		return false, "", businessKey, state, fmt.Errorf("fog_hub is not string")
	}
	return true, fogHubId, businessKey, state, nil
}
