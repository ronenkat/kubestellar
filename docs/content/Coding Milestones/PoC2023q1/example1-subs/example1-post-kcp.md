<!--example1-post-kcp-start-->
### Use the root workspace.

Use the following commands.

```shell
kubectl ws root
```

### Get and build or install KubeStellar

Contributors interested in running the Continuous Integration scripts
locally, or from an existing git development branch, should skip the Fetch step
and resume reading at the Build step.

For new users, download and build, or install, <a href="{{config.repo_url}}">KubeStellar</a>,
according to your preference.  That is, either (a) `git clone` the
repo and then `make build` to populate its `bin` directory, or (b)
fetch the binary archive appropriate for your machine from a release
and unpack it (creating a `bin` directory).  The commands exhibited
just below assume that the repo has been fetched but not yet built.

Fetch step:

```{.base}
git clone -b {{ config.ks_branch }} {{ config.repo_url }}
cd kubestellar
```

Build step:

```shell
echo "current path is $PWD"
make build
export PATH=$(pwd)/bin:$PATH
```

In the following exhibited command lines, the commands described as
"KubeStellar commands" and the commands that start with `kubectl
kubestellar` rely on the KubeStellar `bin` directory being on the
`$PATH`.  Alternatively you could invoke them with explicit pathnames.
The kubectl plugin lines use fully specific executables (e.g.,
`kubectl kubestellar prep-for-syncer` corresponds to
`bin/kubectl-kubestellar-prep_for_syncer`).

### Initialize the KubeStellar platform

In this step KubeStellar creates and populates the Edge Service
Provider Workspace (ESPW), which exports the KubeStellar API, and also
augments the `root:compute` workspace from kcp TMC as needed here.
Those augmentation consists of adding authorization to update the
relevant `/status` and `/scale` subresources (missing in kcp TMC) and
extending the supported subset of the Kubernetes API for managing
containerized workloads from the four resources built into kcp TMC
(`Deployment`, `Pod`, `Service`, and `Ingress`) to the other ones that
are namespaced and are meaningful in KubeStellar.

```shell
kubestellar init
```

### Create an inventory management workspace.
```shell
kubectl ws root
kubectl ws create imw-1 
```
### Create SyncTarget and Location objects to represent the florin and guilder clusters

Use the following two commands. They label both florin and guilder
with `env=prod`, and also label guilder with `extended=si`.

```shell
kubectl ws root:imw-1
kubectl kubestellar ensure location florin  loc-name=florin  env=prod
kubectl kubestellar ensure location guilder loc-name=guilder env=prod extended=si
echo "decribe the florin location object"
kubectl describe location.edge.kcp.io florin
```

Those two script invocations are equivalent to creating the following
four objects.

```yaml
apiVersion: edge.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: florin
  labels:
    id: florin
    loc-name: florin
    env: prod
---
apiVersion: edge.kcp.io/v1alpha1
kind: Location
metadata:
  name: florin
  labels:
    loc-name: florin
    env: prod
spec:
  resource: {group: edge.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: florin}
---
apiVersion: edge.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: guilder
  labels:
    id: guilder
    loc-name: guilder
    env: prod
    extended: si
---
apiVersion: edge.kcp.io/v1alpha1
kind: Location
metadata:
  name: guilder
  labels:
    loc-name: guilder
    env: prod
    extended: si
spec:
  resource: {group: edge.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: guilder}
```

That script also deletes the Location named `default`, which is not
used in this PoC, if it shows up.

<!--example1-post-kcp-end-->
