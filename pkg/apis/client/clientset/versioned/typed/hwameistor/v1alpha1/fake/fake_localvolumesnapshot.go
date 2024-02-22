// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeLocalVolumeSnapshots implements LocalVolumeSnapshotInterface
type FakeLocalVolumeSnapshots struct {
	Fake *FakeHwameistorV1alpha1
	ns   string
}

var localvolumesnapshotsResource = schema.GroupVersionResource{Group: "hwameistor.io", Version: "v1alpha1", Resource: "localvolumesnapshots"}

var localvolumesnapshotsKind = schema.GroupVersionKind{Group: "hwameistor.io", Version: "v1alpha1", Kind: "LocalVolumeSnapshot"}

// Get takes name of the localVolumeSnapshot, and returns the corresponding localVolumeSnapshot object, and an error if there is any.
func (c *FakeLocalVolumeSnapshots) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.LocalVolumeSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(localvolumesnapshotsResource, c.ns, name), &v1alpha1.LocalVolumeSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LocalVolumeSnapshot), err
}

// List takes label and field selectors, and returns the list of LocalVolumeSnapshots that match those selectors.
func (c *FakeLocalVolumeSnapshots) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.LocalVolumeSnapshotList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(localvolumesnapshotsResource, localvolumesnapshotsKind, c.ns, opts), &v1alpha1.LocalVolumeSnapshotList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.LocalVolumeSnapshotList{ListMeta: obj.(*v1alpha1.LocalVolumeSnapshotList).ListMeta}
	for _, item := range obj.(*v1alpha1.LocalVolumeSnapshotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested localVolumeSnapshots.
func (c *FakeLocalVolumeSnapshots) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(localvolumesnapshotsResource, c.ns, opts))

}

// Create takes the representation of a localVolumeSnapshot and creates it.  Returns the server's representation of the localVolumeSnapshot, and an error, if there is any.
func (c *FakeLocalVolumeSnapshots) Create(ctx context.Context, localVolumeSnapshot *v1alpha1.LocalVolumeSnapshot, opts v1.CreateOptions) (result *v1alpha1.LocalVolumeSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(localvolumesnapshotsResource, c.ns, localVolumeSnapshot), &v1alpha1.LocalVolumeSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LocalVolumeSnapshot), err
}

// Update takes the representation of a localVolumeSnapshot and updates it. Returns the server's representation of the localVolumeSnapshot, and an error, if there is any.
func (c *FakeLocalVolumeSnapshots) Update(ctx context.Context, localVolumeSnapshot *v1alpha1.LocalVolumeSnapshot, opts v1.UpdateOptions) (result *v1alpha1.LocalVolumeSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(localvolumesnapshotsResource, c.ns, localVolumeSnapshot), &v1alpha1.LocalVolumeSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LocalVolumeSnapshot), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeLocalVolumeSnapshots) UpdateStatus(ctx context.Context, localVolumeSnapshot *v1alpha1.LocalVolumeSnapshot, opts v1.UpdateOptions) (*v1alpha1.LocalVolumeSnapshot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(localvolumesnapshotsResource, "status", c.ns, localVolumeSnapshot), &v1alpha1.LocalVolumeSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LocalVolumeSnapshot), err
}

// Delete takes name of the localVolumeSnapshot and deletes it. Returns an error if one occurs.
func (c *FakeLocalVolumeSnapshots) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(localvolumesnapshotsResource, c.ns, name), &v1alpha1.LocalVolumeSnapshot{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeLocalVolumeSnapshots) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(localvolumesnapshotsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.LocalVolumeSnapshotList{})
	return err
}

// Patch applies the patch and returns the patched localVolumeSnapshot.
func (c *FakeLocalVolumeSnapshots) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LocalVolumeSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(localvolumesnapshotsResource, c.ns, name, pt, data, subresources...), &v1alpha1.LocalVolumeSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.LocalVolumeSnapshot), err
}