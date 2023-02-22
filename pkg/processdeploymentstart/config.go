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

type Config struct {
	ProcessEngineWrapperUrl     string `json:"process_engine_wrapper_url"`
	FogProcessDeploymentUrl     string `json:"fog_process_deployment_url"`
	WorkerParamPrefix           string `json:"worker_param_prefix"`
	ProcessDeploymentModuleType string `json:"process_deployment_module_type"`
}
