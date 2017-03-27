# oVirt-Containers 4.1
The repository includes image-specifications (for docker currently) and yaml
manifests for openshift to run oVirt deployment (oVirt-Engine and oVirt-Node).

## Pre-requisites
Must use oc tool version 1.5.0 - https://github.com/openshift/origin/releases

| WARNING |
| ---- |
| origin-clients rpm installation adds to /bin oc binary that might be older - verify that you work with 1.5 by "oc version" |

## Getting openshift environment
There are two options - running a cluster of openshift locally or using
Minishift VM:
### Orchestration using Minishift
Minisift is a VM running openshift cluster in it. This mostly being used for
testing for easy and quicker deployment without changing local environemnt
- Install minishift - https://github.com/minishift/minishift
- Run the following

```
export OCTAG=v1.5.0-rc.0
export PROJECT=ovirt
export LATEST_MINISHIFT_CENTOS_ISO_BASE=$(curl -I https://github.com/minishift/minishift-centos-iso/releases/latest | grep "Location" | cut -d: -f2- | tr -d '\r' | xargs)
export MINISHIFT_CENTOS_ISO=${LATEST_MINISHIFT_CENTOS_ISO_BASE/tag/download}/minishift-centos7.iso

minishift start --memory 4096 --iso-url=$MINISHIFT_CENTOS_ISO --openshift-version=$OCTAG
export PATH=$PATH:~/.minishift/cache/oc/$OCTAG
```
### Orchestration using 'oc cluster up'
Just follow https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md#linux

## Load oVirt to openshift instructions
### Login to openshift as system admin
```
oc login -u system:admin
```

### Create oVirt project
```
export PROJECT=ovirt
oc new-project $PROJECT --description="oVirt" --display-name="oVirt"
```

### Add administrator permissions to 'developer' user account
```
oc adm policy add-role-to-user admin developer -n $PROJECT
```

### Force a permissive security context constraints
Allows the usage of root account inside engine pod
```
oc create serviceaccount useroot
oc adm policy add-scc-to-user anyuid -z useroot
```

### Allows host advanced privileges inside the node pod
```
oc create serviceaccount privilegeduser
oc adm policy add-scc-to-user privileged -z privilegeduser
```

### Create engine and node deployments and add them to the project
Please note that the engine deployment is configured as paused
```
oc create -f os-manifests -R
```

#### To deploy just engine
```
oc create -f os-manifests/engine -R
```

#### To deploy just node
```
oc create -f os-manifests/node -R
```

### Change the hostname for the deployments
According to the hostname that was assigned to the associated route
```
oc set env dc/ovirt-engine -c ovirt-engine OVIRT_FQDN=$(oc describe routes ovirt-engine | grep "Requested Host:" | cut -d: -f2 | xargs)
oc set env ds/vdsm-kube-ds -c vdsm-kube ENGINE_FQDN=$(oc describe routes ovirt-engine | grep "Requested Host:" | cut -d: -f2 | xargs)
```

### Unpause deployments
```
oc patch dc/ovirt-engine --patch '{"spec":{"paused": false}}'
oc patch ds/vdsm-kube-ds --patch '{"spec":{"paused": false}}'
```

Now you should be able to login as developer user (developer:admin) to the
$PROJECT project, the server is accessible via web console at
$(minishift console --url)" or locally at https://localhost:8443.
