## Pre-requisites

Get kubebuilder stable release:

```
curl -L -O https://github.com/kubernetes-sigs/kubebuilder/releases/download/v1.0.1/kubebuilder_1.0.1_linux_amd64.tar.gz
```

Install go and prepare your gopath and project:

```
export GOPATH=$HOME/projects/go
mkdir -p $GOPATH/src/github.com/feloy/operator && cd $_
```

Install dep:
```
go get -u github.com/golang/dep/cmd/dep
```

Add $HOME/projects/go/bin to your PATH

## Create the project

Initialize the project:

```
kubebuilder init --domain anevia.com --license none --owner Anevia
```

At this point, you get a project template with a `Makefile`, a `Dockerfile` a basic `manager` and some default yaml files.

## Create a custom resource

```
kubebuilder create api --group cluster --version v1 --kind CdnCluster
```
This will create a resource under `pkg/apis` and an operator under `pkg/controller`.

The created files are:

- the **generated** CRD in yaml format:
  ```
  config/crds/cluster_v1_cdncluster.yaml
  ```

- the **generated** role and binding necessary for operator execution in the cluster:
  ```
  config/rbac/rbac_role.yaml
  config/rbac/rbac_role_binding.yaml
  ```

- a **generated** sample custom resource
  ```
  config/samples/cluster_v1_cdncluster.yaml
  ```

- the **sources** for the new custom resource:
  ```
  pkg/apis/
  ├ addtoscheme_cluster_v1.go
  ├ apis.go
  └ cluster
   ├ group.go
   └ v1
     ├ cdncluster_types.go # the structure definition
     ├ cdncluster_types_test.go # testing the structure
     ├ doc.go
     ├ register.go
     ├ v1_suite_test.go
     └ zz_generated.deepcopy.go
  ```

- the **sources** for the operator:
  ```
  pkg/controller/
  ├ add_cdncluster.go
  ├ cdncluster
  │ ├ cdncluster_controller.go # the reconcile function
  │ ├ cdncluster_controller_suite_test.go
  │ └ cdncluster_controller_test.go # testing the reconcile func
  └ controller.go
  ```

## Deploying the sample Custom resource definition

Verify no CRD is deployed:
```
kubectl get crd
```

Deploy CRD:
```
make install
error: error validating "config/crds/cluster_v1_cdncluster.yaml": error validating data: [ValidationError(CustomResourceDefinition.status): missing required field "conditions" in io.k8s.apiextensions-apiserver.pkg.apis.apiextensions.v1beta1.CustomResourceDefinitionStatus, ValidationError(CustomResourceDefinition.status): missing required field "storedVersions" in io.k8s.apiextensions-apiserver.pkg.apis.apiextensions.v1beta1.CustomResourceDefinitionStatus]; if you choose to ignore these errors, turn validation off with --validate=false
```

Issue #339 (https://github.com/kubernetes-sigs/kubebuilder/issues/339)

Patch your Makefile:

```
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all
+	sed -i 's/  conditions: null/  conditions: []/' config/crds/cluster_v1_cdncluster.yaml
+	echo '  storedVersions: []' >> config/crds/cluster_v1_cdncluster.yaml
```

```
kubectl get crd
cdnclusters.cluster.anevia.com   5s
```

Create a new instance of the custom resource with the provided sample:
```
$ kubectl get cdncluster.cluster.anevia.com
No resources found.
$ kubectl apply -f config/samples/cluster_v1_cdncluster.yaml
cdncluster.cluster.anevia.com "cdncluster-sample" created
$ kubectl get cdncluster.cluster.anevia.com
NAME                AGE
cdncluster-sample   5s
```

You can now delete it:
```
$ kubectl delete cdncluster.cluster.anevia.com cdncluster-sample
cdncluster.cluster.anevia.com "cdncluster-sample" deleted
```

## Customizing the custom resource definition

You can customize the CRD by editing the file `pkg/apis/cluster/v1/cdncluster_types.go`.

The specs part is editable in the `CdnClusterSpec` structure while the status part is editable in the `CdnClusterStatus` one.

Let's add a `Role` field in the specs, and a `State` field in the status:

```
// CdnClusterSpec defines the desired state of CdnCluster
type CdnClusterSpec struct {
    // Role of the CDN cluster, can be 'balancer' or 'cache'
    Role string `json:"role"`
}

// CdnClusterStatus defines the observed state of CdnCluster
type CdnClusterStatus struct {
    // State of the CDN cluster
    State string `json:"state"`
}
```

Note that fields must have json tags.

You can re-generate the yaml files used to deploy the CRD, and examine the differences:
```
$ make manifests
$ git diff config/crds/cluster_v1_cdncluster.yaml 
diff --git a/config/crds/cluster_v1_cdncluster.yaml b/config/crds/cluster_v1_cdncluster.yaml
index 8d0dcbb..fe0efaf 100644
--- a/config/crds/cluster_v1_cdncluster.yaml
+++ b/config/crds/cluster_v1_cdncluster.yaml
@@ -21,8 +21,18 @@ spec:
         metadata:
           type: object
         spec:
+          properties:
+            role:
+              type: string
+          required:
+          - role
           type: object
         status:
+          properties:
+            state:
+              type: string
+          required:
+          - state
           type: object
       type: object
   version: v1
```

You can see that the `role` and `state` properties have been added to the definition of the CRD, and are marked as **required**.

## Making a field not required

If you want a field to be not required, you can use the `omitempty` flag in the json tag associated with this field:
```
// CdnClusterStatus defines the observed state of CdnCluster
type CdnClusterStatus struct {
    State string `json:"state,omitempty"`
}
```

then re-generate the manifests again:
```
$ make manifests
$ git diff config/crds/cluster_v1_cdncluster.yaml 
diff --git a/config/crds/cluster_v1_cdncluster.yaml b/config/crds/cluster_v1_cdncluster.yaml
index fe0efaf..f663eba 100644
--- a/config/crds/cluster_v1_cdncluster.yaml
+++ b/config/crds/cluster_v1_cdncluster.yaml
@@ -31,8 +31,6 @@ spec:
           properties:
             state:
               type: string
-          required:
-          - state
           type: object
       type: object
   version: v1
```

The `state` field is not required anymore.

## Completing the Custom resource definition

We want our CDN clusters to redirect requests to *source clusters* depending on some condition on the path of the requested URL. For this, we add a list of `sources` to the definition of a CDN cluster and a source is defined by the name of the source CDN cluster and the path condition to redirect to this cluster.

The list of sources cannot be omitted (but can be an empty array), and a path condition can be omitted, in the case of a default source cluster (the one selected if no other path condition in other sources matches):
```
// CdnClusterSource defines a source cluster of a cluster
type CdnClusterSource struct {
    // The name of the source cluster
    Name string `json:"name"`
    // The path condition to enter this cluster,
    // can be omitted for the default source
    PathCondition string `json:"pathCondition,omitempty"`
}

// CdnClusterSpec defines the desired state of CdnCluster
type CdnClusterSpec struct {
    // Role must be 'balancer' or 'cache'
    Role string `json:"role"`
    // Sources is the list of source clusters for this cluster
    Sources []CdnClusterSource `json:"sources"`
}
```

## Creating sample custom resource instances

Here we create three instances of CDN clusters. A first instance of balancers, which will have two sources, one cluster of caches for Live requests and another for VOD requests:

```
apiVersion: cluster.anevia.com/v1
kind: CdnCluster
metadata:
  name: balancer
spec:
  role: balancer
  sources:
  - name: cache-live
    pathCondition: ^/live/
  - name: cache-vod
    pathCondition: ^/vod/

---

apiVersion: cluster.anevia.com/v1
kind: CdnCluster
metadata:
  name: cache-live
spec:
  role: cache
  sources: []

---

apiVersion: cluster.anevia.com/v1
kind: CdnCluster
metadata:
  name: cache-vod
spec:
  role: cache
  sources: []
```

To deploy the instances: 
```
$ kubectl apply -f config/samples/cluster_v1_cdncluster.yaml 
cdncluster.cluster.anevia.com "balancer" created
cdncluster.cluster.anevia.com "cache-live" created
cdncluster.cluster.anevia.com "cache-vod" created
```

## Testing the creation of CdnCluster instances

```
$ make test
... spec.sources in body must be of type array: "null" ...
```

In the tests, we create a CDN cluster with this command:

```
created := &CdnCluster{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}}
```

In Go, an omitted field in a struct is equivalent to its zero value, so the command is equivalent to:
```
created := &CdnCluster{
  ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
  Spec: CdnClusterSpec{
    Role: "",
    Sources: nil,
  },
}
```

The Kubernetes API does not accept a nil value for the Sources with an array type; you have to define the sources with an empty array, for example:
```
created := &CdnCluster{
  ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
  Spec: CdnClusterSpec{
    Role: "",
    Sources: []CdnClusterSource{},
  },
}
```
or with a more complete specification:
```
created := &CdnCluster{
    ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
    Spec: CdnClusterSpec{
        Role: "balancer",
        Sources: []CdnClusterSource{
            {
                Name:          "cache-live",
                PathCondition: "^/live/",
            },
            {
                Name:          "cache-vod",
                PathCondition: "^/vod/",
            },
        },
    },
}
  ```

This time, the tests should pass:
```
$ make test
ok  	operator/pkg/apis/cluster/v1
```

## Generating the clienset for the CRD

At this time, you can create new CDN clusters with the `kubectl` command. If you want to create new CDN clusters
from a Go application, you will need a specific clientset for this resource.

Install code-generator and select the branch for Kubernetes 1.10:
```
$ go get k8s.io/code-generator
$ cd $GOPATH/src/k8s.io/code-generator/
$ git checkout -b 1.10 origin/release-1.10
```

Run the code-generator to generate the clientset:
```
$ cd $GOPATH
$ ./src/k8s.io/code-generator/generate-groups.sh \
    client \
    github.com/feloy/operator \
    github.com/feloy/operator/pkg/apis \
    cluster:v1
```

Remove the `-zz_generated.*` entry from `.gitignore` so the generated deepcopy file is added to the repository.

Add a missing declaration of `AddToScheme` to the `pkg/apis/cluster/v1/register.go` file:
```
diff --git a/pkg/apis/cluster/v1/register.go b/pkg/apis/cluster/v1/register.go
index bfb6952..9e9086c 100644
--- a/pkg/apis/cluster/v1/register.go
+++ b/pkg/apis/cluster/v1/register.go
@@ -23,4 +23,5 @@ var (

        // SchemeBuilder is used to add go types to the GroupVersionKind scheme
        SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
+       AddToScheme   = SchemeBuilder.AddToScheme
 )
```

## Using the generated clientset

You can create a new Go project in your gopath with the following `main.go` file:
```
package main

import (
	"os"
	"path/filepath"

	clientsetCdnclusterv1 "github.com/feloy/operator/clientset/versioned"
	cdnclusterv1 "github.com/feloy/operator/pkg/apis/cluster/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, _ := clientsetCdnclusterv1.NewForConfig(config)
	created := &cdnclusterv1.CdnCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
		Spec: cdnclusterv1.CdnClusterSpec{
			Role: "balancer",
			Sources: []cdnclusterv1.CdnClusterSource{
				{
					Name:          "cache-live",
					PathCondition: "^/live/",
				},
				{
					Name:          "cache-vod",
					PathCondition: "^/vod/",
				},
			},
		},
	}
	clientset.ClusterV1().CdnClusters("default").Create(created)
}
```

You will need to use the correct `apimachinery` and `client-go` versions, compatible with the versions used by the kubebuilder tool, with this `Gopkg.toml` file:
```
[[override]]
  name = "k8s.io/apimachinery"
  version = "kubernetes-1.10.0"

[[override]]
  name = "k8s.io/client-go"
  version="kubernetes-1.10.1"
```
