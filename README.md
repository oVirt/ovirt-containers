# oVirt-Containers - master snapshot


> IMPORTANT: This project has been dropped from oVirt
>
> Keeping the repo only for reference.



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
Minishift is a VM running openshift cluster in it. This mostly being used for
testing for easy and quicker deployment without changing local environemnt
- Install minishift - https://github.com/minishift/minishift
  #### Nested virtualization support for the minishift VM
  Since we are going to run VMs inside the minishift VM, it will need nested virtualization support. Minishift documentation suggests to install docker-machine-driver-kvm v0.7.0 but this is not enough to get virtualization support and a more recent one is needed; the right instructions (assuming centos7) to get it are:
  ```
  curl -L https://github.com/docker/machine/releases/download/v0.11.0/docker-machine-`uname -s`-`uname -m` >/tmp/docker-machine &&
    chmod +x /tmp/docker-machine &&
    sudo cp /tmp/docker-machine /usr/local/bin/docker-machine
  curl -L https://github.com/dhiltgen/docker-machine-kvm/releases/download/v0.10.0/docker-machine-driver-kvm-centos7 > /usr/local/bin/docker-machine-driver-kvm &&
    chmod +x /usr/local/bin/docker-machine-driver-kvm
  ```
  Nested virtualization support is required on your physical system, please check it with:
  ```
  cat /sys/module/kvm_intel/parameters/nested
  Y
  ```
  If nested virtualization is not supported, please check your OS documentation about how to enable it.

- Run the following
  ```
  export OCTAG=v1.5.0
  export PROJECT=ovirt
  export LATEST_MINISHIFT_CENTOS_ISO_BASE=$(curl -I https://github.com/minishift/minishift-centos-iso/releases/latest | grep "Location" | cut -d: -f2- | tr -d '\r' | xargs)
  export MINISHIFT_CENTOS_ISO=${LATEST_MINISHIFT_CENTOS_ISO_BASE/tag/download}/minishift-centos7.iso

  minishift start --memory 6144 --cpus 4 --iso-url=$MINISHIFT_CENTOS_ISO --openshift-version=$OCTAG
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

### Allows host advanced privileges inside the vdsc pod
```
oc create serviceaccount privilegeduser
oc adm policy add-scc-to-user privileged -z privilegeduser
```

### Create engine and vdsc deployments and add them to the project
Please note that the engine deployment is configured as paused
```
oc create -f os-manifests -R
```

#### To deploy just engine
```
oc create -f os-manifests/engine -R
```

#### To deploy just vdsc
```
oc create -f os-manifests/vdsc -R
```

### Change the hostname for engine deployment and unpause it
According to the hostname that was assigned to the associated route
```
oc set env dc/ovirt-engine -c ovirt-engine OVIRT_FQDN=$(oc describe routes ovirt-engine | grep "Requested Host:" | cut -d: -f2 | xargs)
oc set env dc/ovirt-engine -c ovirt-engine SPICE_PROXY=http://$(oc describe routes ovirt-spice-proxy | grep "Requested Host:" | cut -d: -f2 | xargs):3128
oc patch dc/ovirt-engine --patch '{"spec":{"paused": false}}'
```

Now you should be able to login as developer user (developer:admin) to the
$PROJECT project, the server is accessible via web console at
$(minishift console --url)" or locally at https://localhost:8443.
