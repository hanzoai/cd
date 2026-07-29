package application

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
	appclientset "github.com/hanzoai/cd/pkg/client/clientset/versioned"
	appv1alpha1 "github.com/hanzoai/cd/pkg/client/clientset/versioned/typed/application/v1alpha1"
	applisters "github.com/hanzoai/cd/pkg/client/listers/application/v1alpha1"
	"github.com/hanzoai/cd/util"

	"k8s.io/apimachinery/pkg/labels"
)

// deepCopyApplicationLister wraps an ApplicationLister and returns deep copies of the applications.
type deepCopyApplicationLister struct {
	applisters.ApplicationLister
}

// List lists all Applications in the indexer and returns deep copies.
func (d *deepCopyApplicationLister) List(selector labels.Selector) ([]*v1alpha1.Application, error) {
	apps, err := d.ApplicationLister.List(selector)
	if err != nil {
		return nil, err
	}
	deepCopiedApps := util.SliceCopy(apps)
	return deepCopiedApps, nil
}

// Applications return an object that can list and get Applications and returns deep copies.
func (d *deepCopyApplicationLister) Applications(namespace string) applisters.ApplicationNamespaceLister {
	return &deepCopyApplicationNamespaceLister{
		ApplicationNamespaceLister: d.ApplicationLister.Applications(namespace),
	}
}

// deepCopyApplicationNamespaceLister wraps an ApplicationNamespaceLister and returns deep copies of the applications.
type deepCopyApplicationNamespaceLister struct {
	applisters.ApplicationNamespaceLister
}

// List lists all Applications in the indexer for a given namespace and returns deep copies.
func (d *deepCopyApplicationNamespaceLister) List(selector labels.Selector) ([]*v1alpha1.Application, error) {
	apps, err := d.ApplicationNamespaceLister.List(selector)
	if err != nil {
		return nil, err
	}
	deepCopiedApps := util.SliceCopy(apps)
	return deepCopiedApps, nil
}

// Get retrieves the Application from the indexer for a given namespace and name and returns a deep copy.
func (d *deepCopyApplicationNamespaceLister) Get(name string) (*v1alpha1.Application, error) {
	app, err := d.ApplicationNamespaceLister.Get(name)
	if err != nil {
		return nil, err
	}
	return app.DeepCopy(), nil
}

type deepCopyAppClientset struct {
	appclientset.Interface
}

func (d *deepCopyAppClientset) AppV1alpha1() appv1alpha1.AppV1alpha1Interface {
	return &deepCopyAppV1alpha1Client{
		AppV1alpha1Interface: d.Interface.AppV1alpha1(),
	}
}

// GetUnderlyingClientSet returns the underlying clientset.Interface.
// Unit tests should only call this
func (d *deepCopyAppClientset) GetUnderlyingClientSet() appclientset.Interface {
	return d.Interface
}

type deepCopyAppV1alpha1Client struct {
	appv1alpha1.AppV1alpha1Interface
}

func (d *deepCopyAppV1alpha1Client) RESTClient() rest.Interface {
	return d.AppV1alpha1Interface.RESTClient()
}

func (d *deepCopyAppV1alpha1Client) AppProjects(namespace string) appv1alpha1.AppProjectInterface {
	return &deepCopyAppProjectClient{
		AppProjectInterface: d.AppV1alpha1Interface.AppProjects(namespace),
	}
}

func (d *deepCopyAppV1alpha1Client) ApplicationSets(namespace string) appv1alpha1.ApplicationSetInterface {
	return &deepCopyApplicationSetClient{
		ApplicationSetInterface: d.AppV1alpha1Interface.ApplicationSets(namespace),
	}
}

func (d *deepCopyAppV1alpha1Client) Applications(namespace string) appv1alpha1.ApplicationInterface {
	return &deepCopyApplicationClient{
		ApplicationInterface: d.AppV1alpha1Interface.Applications(namespace),
	}
}

type deepCopyApplicationClient struct {
	appv1alpha1.ApplicationInterface
}

func (d *deepCopyApplicationClient) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Application, error) {
	app, err := d.ApplicationInterface.Get(ctx, name, options)
	if err != nil {
		return nil, err
	}
	return app.DeepCopy(), nil
}

func (d *deepCopyApplicationClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.ApplicationList, error) {
	appList, err := d.ApplicationInterface.List(ctx, opts)
	if err != nil {
		return nil, err
	}
	return appList.DeepCopy(), nil
}

type deepCopyAppProjectClient struct {
	appv1alpha1.AppProjectInterface
}

func (d *deepCopyAppProjectClient) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.AppProject, error) {
	appProject, err := d.AppProjectInterface.Get(ctx, name, options)
	if err != nil {
		return nil, err
	}
	return appProject.DeepCopy(), nil
}

func (d *deepCopyAppProjectClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.AppProjectList, error) {
	appProjectList, err := d.AppProjectInterface.List(ctx, opts)
	if err != nil {
		return nil, err
	}
	return appProjectList.DeepCopy(), nil
}

type deepCopyApplicationSetClient struct {
	appv1alpha1.ApplicationSetInterface
}

func (d *deepCopyApplicationSetClient) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.ApplicationSet, error) {
	appSet, err := d.ApplicationSetInterface.Get(ctx, name, options)
	if err != nil {
		return nil, err
	}
	return appSet.DeepCopy(), nil
}

func (d *deepCopyApplicationSetClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.ApplicationSetList, error) {
	appSetList, err := d.ApplicationSetInterface.List(ctx, opts)
	if err != nil {
		return nil, err
	}
	return appSetList.DeepCopy(), nil
}
