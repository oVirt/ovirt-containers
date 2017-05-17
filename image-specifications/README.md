# Building oVirt Docker Images
This document provides instructions for building local versions of the Docker
images that are used by an ovirt-containers deployment.  Unless you need to
make changes you can skip these instructions.  By default, deployment will pull
all of the required images from DockerHub.

## Build
Start working on your development machine within an appropriate checkout of
this repository.  If you are using minishift, set up your environment to use
the minishift docker instance:
 ```
eval $( minishift docker-env )
 ```

To automatically rebuild all required Docker images:
```
make build
```

#### Manual build instructions

Check the openshift manifest files
(`os-manifests/*/*-deployment.yaml`) to determine which docker images you need.
At the time of this writing we require the following images:
 - ovirt/vdsc-syslog:master
 - ovirt/vdsc:master
 - ovirt/engine-spice-proxy:master
 - ovirt/engine-database:master
 - ovirt/engine:master

Look in the Dockerfile associated with each image (eg.
`image-specifications/vdsc/Dockerfile`) and discover that the images depend
on the base layer `ovirt/base:master`.

We are not concerned with locally building images outside of the ovirt
namespace (eg. CentOS) and will allow these to be pulled from DockerHub.

Build the images (make sure to tag each one properly using '-t'):
```
docker build -t ovirt/base:master image-specifications/base
docker build -t ovirt/vdsc-syslog:master image-specifications/vdsc-syslog
docker build -t ovirt/vdsc:master image-specifications/vdsc
docker build -t ovirt/engine-spice-proxy:master image-specifications/engine-spice-proxy
docker build -t ovirt/engine-database:master image-specifications/engine-database
docker build -t ovirt/engine:master image-specifications/engine
```

After building make sure the updated images are available to the Openshift
docker instance.  Unless you targeted the minishift docker environment as
described above you may need to transfer the images.  This can be done with
docker save/load:
```
docker save ovirt/vdsc:master | ssh <oc node> 'docker load'
```

## Install
After making changes to the docker images or openshift deployment files you
may want to replace an existing deployment.  Use a similar command to delete
the application as you used to create it, for example:
```
oc delete -f os-manifests -R
```

Then create the application again to use the updated images and manifests:
```
oc create -f os-manifests -R
```
