/*
Copyright 2018 Anevia.
*/

package cdncluster

import (
	"testing"
	"time"

	clusterv1 "github.com/feloy/operator/pkg/apis/cluster/v1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

/*
func TestReconcile(t *testing.T) {

	var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
	var depKey = types.NamespacedName{Name: "foo-deployment", Namespace: "default"}

	const timeout = time.Second * 5

	g := gomega.NewGomegaWithT(t)

	instance := &clusterv1.CdnCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: clusterv1.CdnClusterSpec{
			Sources: []clusterv1.CdnClusterSource{},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	// Create the CdnCluster object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
		Should(gomega.Succeed())

	// Delete the Deployment and expect Reconcile to be called for Deployment deletion
	g.Expect(c.Delete(context.TODO(), deploy)).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
		Should(gomega.Succeed())

	// Manually delete Deployment since GC isn't enabled in the test control plane
	g.Expect(c.Delete(context.TODO(), deploy)).To(gomega.Succeed())

}
*/

func TestReconcile2(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup the Manager and Controller.
	// Wrap the Controller Reconcile function
	// so it writes each request to a channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	instance := &clusterv1.CdnCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "default",
		},
		Spec: clusterv1.CdnClusterSpec{
			Sources: []clusterv1.CdnClusterSource{},
		},
	}

	// Create the CdnCluster object
	// and expect the Reconcile to be called
	// with the instance namespace and name as parameter
	err = c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)

	var expectedRequest = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "foo2",
			Namespace: "default",
		},
	}
	const timeout = time.Second * 5

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// Expect that a Deployment is created
	deploy := &appsv1.Deployment{}
	var depKey = types.NamespacedName{
		Name:      "foo2-deployment",
		Namespace: "default",
	}
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).Should(gomega.Succeed())

	// Delete the Deployment and expect Reconcile
	// to be called for Deployment deletion
	// and Deployment to be created again
	g.Expect(c.Delete(context.TODO(), deploy)).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).Should(gomega.Succeed())
}
