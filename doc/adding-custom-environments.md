## Adding custom environments

Add custom environment loaded in all 3scale products.

Here is an example of a policy that is loaded in all services: `custom_env.lua`

```lua
local cjson = require('cjson')
local PolicyChain = require('apicast.policy_chain')
local policy_chain = context.policy_chain

local logging_policy_config = cjson.decode([[
{
  "enable_access_logs": false,
  "custom_logging": "\"{{request}}\" to service {{service.id}} and {{service.name}}"
}
]])

policy_chain:insert( PolicyChain.load_policy('logging', 'builtin', logging_policy_config), 1)

return {
  policy_chain = policy_chain,
  port = { metrics = 9421 },
}
```

### Prerequisites

* One or more custom environment in lua code.

### Adding custom environment

#### Create secret with the custom environment content

```
oc create secret generic custom-env-1 --from-file=./env11.lua
```

**NOTE**: a secret can host multiple custom environments. The operator will load each one of them.

#### Configure and deploy APIcast CR with the custom environment

`apicast.yaml` content (only relevant content shown):

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  customEnvironments:
    - secretRef:
        name: custom-env-1
```

**NOTE**: Multiple custom environment secrets can be added. The operator will load each one of them.

```
oc apply -f apicast.yaml
```

The APIcast custom resource allows adding multiple custom policies.

**NOTE**: If secret does not exist, the operator would mark the custom resource as failed. The Deployment object would fail if secret does not exist.

*NOTE*: Once apicast has been deployed, the content of the secret should not be updated externally.
If the content of the secret is updated externally, after apicast has been deployed, the container can automatically see the changes.
However, apicast has the environment already loaded and it does not change the behavior.

If the custom environment content needs to be changed, there are two options:

* [**recommended way**] Create another secret with a different name and update the APIcast custom resource field `spec.customEnvironments[].secretRef.name`. The operator will trigger a rolling update loading the new custom environment content.
* Update the existing secret content and redeploy apicast turning `spec.replicas` to 0 and then back to the previous value.
