/*
Copyright 2022 Contributors to The HwameiStor project.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	scheme "github.com/hwameistor/local-storage/pkg/apis/client/clientset/versioned/scheme"
	v1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// LocalVolumesGetter has a method to return a LocalVolumeInterface.
// A group's client should implement this interface.
type LocalVolumesGetter interface {
	LocalVolumes() LocalVolumeInterface
}

// LocalVolumeInterface has methods to work with LocalVolume resources.
type LocalVolumeInterface interface {
	Create(ctx context.Context, localVolume *v1alpha1.LocalVolume, opts v1.CreateOptions) (*v1alpha1.LocalVolume, error)
	Update(ctx context.Context, localVolume *v1alpha1.LocalVolume, opts v1.UpdateOptions) (*v1alpha1.LocalVolume, error)
	UpdateStatus(ctx context.Context, localVolume *v1alpha1.LocalVolume, opts v1.UpdateOptions) (*v1alpha1.LocalVolume, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.LocalVolume, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.LocalVolumeList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LocalVolume, err error)
	LocalVolumeExpansion
}

// localVolumes implements LocalVolumeInterface
type localVolumes struct {
	client rest.Interface
}

// newLocalVolumes returns a LocalVolumes
func newLocalVolumes(c *HwameistorV1alpha1Client) *localVolumes {
	return &localVolumes{
		client: c.RESTClient(),
	}
}

// Get takes name of the localVolume, and returns the corresponding localVolume object, and an error if there is any.
func (c *localVolumes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.LocalVolume, err error) {
	result = &v1alpha1.LocalVolume{}
	err = c.client.Get().
		Resource("localvolumes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of LocalVolumes that match those selectors.
func (c *localVolumes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.LocalVolumeList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.LocalVolumeList{}
	err = c.client.Get().
		Resource("localvolumes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested localVolumes.
func (c *localVolumes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("localvolumes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a localVolume and creates it.  Returns the server's representation of the localVolume, and an error, if there is any.
func (c *localVolumes) Create(ctx context.Context, localVolume *v1alpha1.LocalVolume, opts v1.CreateOptions) (result *v1alpha1.LocalVolume, err error) {
	result = &v1alpha1.LocalVolume{}
	err = c.client.Post().
		Resource("localvolumes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(localVolume).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a localVolume and updates it. Returns the server's representation of the localVolume, and an error, if there is any.
func (c *localVolumes) Update(ctx context.Context, localVolume *v1alpha1.LocalVolume, opts v1.UpdateOptions) (result *v1alpha1.LocalVolume, err error) {
	result = &v1alpha1.LocalVolume{}
	err = c.client.Put().
		Resource("localvolumes").
		Name(localVolume.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(localVolume).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *localVolumes) UpdateStatus(ctx context.Context, localVolume *v1alpha1.LocalVolume, opts v1.UpdateOptions) (result *v1alpha1.LocalVolume, err error) {
	result = &v1alpha1.LocalVolume{}
	err = c.client.Put().
		Resource("localvolumes").
		Name(localVolume.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(localVolume).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the localVolume and deletes it. Returns an error if one occurs.
func (c *localVolumes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("localvolumes").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *localVolumes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("localvolumes").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched localVolume.
func (c *localVolumes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LocalVolume, err error) {
	result = &v1alpha1.LocalVolume{}
	err = c.client.Patch(pt).
		Resource("localvolumes").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
