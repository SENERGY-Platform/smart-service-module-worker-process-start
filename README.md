## Outputs

### Instance-Id

- Desc: id of created instance
- Variable-Name: process_instance_id

## Inputs

### Deployment-Id

- Desc: defines which process-deployment should be de started
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.process_deployment_id`
- Variable-Name-Example: `process_deployment_start.process_deployment_id`
- Value: string

### Start-Input-Variable

- Desc: sets the (json encoded) value of an input variable
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.input.{{name}}`
- Variable-Name-Example: `process_deployment_start.input.foo`
- Value: string