# oVirt-Containers - master snapshot
The repository includes CI to build images and run manifests definitions of ovirt components over openshift environment

# Details
The orchestration includes ovirt-engine deployment and kube-vdsm deamonset.
* For building images run "/bin/sh automation/build-artifacts.sh"
* To load deployment to openshift run "oc create -f os-manifests -R" (for setting openshift cluster follow [1] or [2] to set up a testing instance on minishift)

Please note that currently this project requires at least openshift v1.5.0-rc.0 so, in order to create a minishift instance with it, the right command line is:
 minishift start --openshift-version=v1.5.0-rc.0

[1] https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md#linux
[2] https://github.com/minishift/minishift/blob/master/README.md#installation

