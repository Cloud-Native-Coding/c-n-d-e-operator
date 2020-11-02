# c-n-d-e Operator

This Operator operates instances of Pods that serve as an in-cluster IDE accessible via Web.

This operator...:

- builds and pushes an IDE-image based on the given `Template`
- creates a new Realm, Client and User in a Keycloak instance
- creates all necessary Kubernetes resources to provide a VM like experience (Pods, Ingresses, oauth-proxy and IDE Pods, etc.)

## Installation

- properly setup the following tools:
  - *Kubernetes*
  - *Keycloak* (as an oauth provider)
  - *cert-manager*, optional (for creating e.g. lets-encrypt certificates)
  - *Ingress controller*
  - *external DNS*, optional (for creating DNS entries)
- configure `kustomization.yaml` in folder `config/default`
  - choose a *namespace*
  - choose a *prefix*
  - optional, provide and uncomment *kaniko-secret* if your builder needs a Secret to push an container image
- provide Keycloak credentials in file `oauth-properties`
- change value for ENV `CNDE_OAUTH_URL` (the URL where Keycloak can be accessed) in file `manager-patch.yaml`
- if you want, do a dry run 1st `kustomize build . | kubectl apply --dry-run=server -f -`
- execute `kustomize build . | kubectl apply -f -`

## Stand Alone Usage

This example snippet of a `kustomization.yaml` creates two build environments using ConfigMaps:
(please also refer to folder `config/examples/builder`)

```yaml
...
configMapGenerator:
- name: dev-env-build-k8s-go
  files:
    - builder-k8s-go/.zshrc
    - builder-k8s-go/Dockerfile
- name: dev-env-build-k8s
  files:
    - builder-k8s/.zshrc
    - builder-k8s/Dockerfile
...
```

The above ConfigMaps are used in the two builder that can be reviewed in file `config/examples/builder/builder.yaml`

```yaml
apiVersion: c-n-d-e.kube-platform.dev/v1alpha1
kind: DevEnv
metadata:
  name: thedeep
spec:
  # the builder created before
  builderName: devenv-builder-k8s
  # the cluster role to use for the IDE
  clusterRoleName: system:aggregate-to-view
  # the role to use for the IDE
  roleName: system:aggregate-to-edit
  # both images have to be the same for now
  # the builder tries to push the image to these tags
  configureImg: eu.gcr.io/myusername/dev-env-thedeep:latest
  devEnvImg: eu.gcr.io/myusername/dev-env-thedeep:latest
  # delete volumes if the DevEnv resource is deleted?
  deleteVolumes: true
  # volume size of the Docker DinD Sidecar
  dockerVolumeSize: 10Gi
  # home volume size of the IDE
  homeVolumeSize: 10Gi
  # host name of the Keycloak host
  keycloakHost: keycloak
  # oauth email address of user
  userEmail: norbert@cloud-native-coding.dev
  # the domain for the Ingress resource to create (DevEnv name will be prefixed)
  userEnvDomain: kubeplatform.my.domain.io
```

## Usage with c-n-d-e Controller

If you are using the c-n-d-e Controller together with the c-n-d-e Dashboard the Resources above will be managed automatically
