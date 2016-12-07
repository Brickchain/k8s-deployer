# k8s-deployer
Deploy multiple repositories containing Kubernetes configs with one command.  

## Dependencies
kubectl needs to be installed and configured to connect to the cluster you want to deploy to.

## Namespaces
If the namespace you specify does not exist, it will create it.

## Environment variable substitution
It will run all Kubernetes configs through a template renderer before applying them,  
so you can use Golang template format variables in your Kubernetes configs.

```yaml
someParam: "{{ .SOME_ENV_VARIABLE }}"
``` 

## Config file format
```yaml
---

namespace: "{{ .CI_BUILD_REF_NAME }}"
defaultBranch: master
kubernetesFolder: "k8s/"

updateRepoVar: "CI_UPSTREAM_PROJECT_NAME"
updateRefVar: "CI_UPSTREAM_BUILD_REF"


repositories:
    - name: someservice
      uri: "git@gitlab.com:group/someservice.git"
    - name: otherservice
      uri: "git@gitlab.com:group/otherservice.git"

```

## Usage
```bash
$ k8s-deployer -h
Usage of k8s-deployer:
  -artifact string
        Create YAML with what was deployed
  -clear-state
        Clear the state for this namespace
  -config string
        Config file
  -namespace string
        Namespace
  -redis string
        Redis state DB. Ex: localhost:6379

# deploy repos specified in config.yml and record state in redis and also write an artifact file
$ k8s-deployer -config config.yml -redis localhost:6379 -artifact state.yml

# deploy a previously recorded state to a temporary namespace
$ k8s-deployer -namespace debugging -config state.yml
```
