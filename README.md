# oVirt-Containers - master snapshot
The repository includes CI to build images and run manifests definitions of ovirt components over openshift environment

# Details
The orchestration includes ovirt-engine deployment and kube-vdsm deamonset.
* For building images run "/bin/sh automation/build-artifacts.sh"
* To load deployment to openshift run "oc create -f os-manifests -R" (for setting openshift cluster follow [1] or [2] to set up a testing instance on minishift)

Please note that currently this project requires at least openshift v1.5.0-rc.0 so, in order to create a minishift instance with it, the right sequence is:
 export OCTAG=v1.5.0-rc.0
 export PROJECT=ovirt
 minishift start --openshift-version=$OCTAG
 export PATH=$PATH:~/.minishift/cache/oc/$OCTAG
 oc login -u system:admin
 oc new-project $PROJECT --description="oVirt" --display-name="oVirt"
 oc adm policy add-role-to-user admin developer -n $PROJECT
 oc create serviceaccount useroot
 oc adm policy add-scc-to-user anyuid -z useroot
 oc create -f os-manifests -R
 oc set env dc/ovirt-engine -c ovirt-engine OVIRT_FQDN=$(oc describe routes ovirt-engine | grep "Requested Host:" | cut -d: -f2 | xargs)
 oc patch dc/ovirt-engine --patch '{"spec":{"paused": false}}'

[1] https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md#linux
[2] https://github.com/minishift/minishift/blob/master/README.md#installation

