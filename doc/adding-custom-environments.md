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

By default, content changes in the secret will not be noticed by the apicast operator.
The apicast operator allows monitoring the secret for changes adding the `apicast.apps.3scale.net/watched-by=apicast` label.
With that label in place, when the content of the secret is changed, the operator will get notified.
Then, the operator will rollout apicast deployment to make the changes effective.
The operator will not take *ownership* of the secret in any way.

```
kubectl label secret custom-env-1 apicast.apps.3scale.net/watched-by=apicast
```

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
