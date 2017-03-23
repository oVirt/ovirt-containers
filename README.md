# oVirt-Containers 4.1

# MUST HAVE
Must use oc tool version 1.5.0 - https://github.com/openshift/origin/releases

WARNING: origin-clients rpm installation adds to /bin oc binary that might be older - verify that you work with 1.5 by "oc version"

# Details
The orchestration includes engine deploymentconfig and kube-vdsm deamonset.
* For building images run "/bin/sh automation/build-artifacts.sh"
* To load deployment to openshift run "oc create -f os-manifests -R" (for setting openshift cluster follow [1] or [2] to set up a testing instance on minishift).

[1] https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md#linux
[2] https://github.com/minishift/minishift/blob/master/README.md#installation

# Orchestration using Minishift
- Install minishift - https://github.com/minishift/minishift
- Run the following

 export OCTAG=v1.5.0-rc.0
 export PROJECT=ovirt
 LATEST_MINISHIFT_CENTOS_ISO_BASE=$(curl -I https://github.com/minishift/minishift-centos-iso/releases/latest | grep "Location" | cut -d: -f2- | tr -d '\r' | xargs)
 MINISHIFT_CENTOS_ISO=${LATEST_MINISHIFT_CENTOS_ISO_BASE/tag/download}/minishift-centos7.iso
 minishift start --memory 4096 --iso-url=$MINISHIFT_CENTOS_ISO --openshift-version=$OCTAG
 export PATH=$PATH:~/.minishift/cache/oc/$OCTAG

 # login as system admin
 oc login -u system:admin

 # create ovirt project
 oc new-project $PROJECT --description="oVirt" --display-name="oVirt"

 # add administrator permissions for the ovirt project to the developer user account
 oc adm policy add-role-to-user admin developer -n $PROJECT

 # force a permissive security context constraints
 # allows the usage of root account inside engine pod
 oc create serviceaccount useroot
 oc adm policy add-scc-to-user anyuid -z useroot
 # allows host IPC inside node pod
 oc create serviceaccount privilegeduser
 oc adm policy add-scc-to-user privileged -z privilegeduser

 # create engine and node deployments and add them to the project
 # please note that the engine deployment is configured as paused
 oc create -f os-manifests -R
 #oc create -f os-manifests/engine -R # if you want to deploy just the engine
 #oc create -f os-manifests/node -R # if you want to deploy just the node

 # change the hostname for the ovirt-engine deployment according to the hostname that was assigned to the associated route
 oc set env dc/ovirt-engine -c ovirt-engine OVIRT_FQDN=$(oc describe routes ovirt-engine | grep "Requested Host:" | cut -d: -f2 | xargs)

 # unpause ovirt-engine deployment
 oc patch dc/ovirt-engine --patch '{"spec":{"paused": false}}'

 # provide login info
 echo "Now you can login as developer user to the $PROJECT project, the server is accessible via web console at $(minishift console --url)"
