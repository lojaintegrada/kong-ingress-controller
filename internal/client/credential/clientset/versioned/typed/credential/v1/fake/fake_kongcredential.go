/*
Copyright 2018 The Kong Authors.

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

package fake

import (
	credential_v1 "github.com/kong/ingress-controller/internal/apis/credential/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKongCredentials implements KongCredentialInterface
type FakeKongCredentials struct {
	Fake *FakeConfigurationV1
	ns   string
}

var kongcredentialsResource = schema.GroupVersionResource{Group: "configuration.konghq.com", Version: "v1", Resource: "kongcredentials"}

var kongcredentialsKind = schema.GroupVersionKind{Group: "configuration.konghq.com", Version: "v1", Kind: "KongCredential"}

// Get takes name of the kongCredential, and returns the corresponding kongCredential object, and an error if there is any.
func (c *FakeKongCredentials) Get(name string, options v1.GetOptions) (result *credential_v1.KongCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kongcredentialsResource, c.ns, name), &credential_v1.KongCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*credential_v1.KongCredential), err
}

// List takes label and field selectors, and returns the list of KongCredentials that match those selectors.
func (c *FakeKongCredentials) List(opts v1.ListOptions) (result *credential_v1.KongCredentialList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kongcredentialsResource, kongcredentialsKind, c.ns, opts), &credential_v1.KongCredentialList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &credential_v1.KongCredentialList{}
	for _, item := range obj.(*credential_v1.KongCredentialList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kongCredentials.
func (c *FakeKongCredentials) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kongcredentialsResource, c.ns, opts))

}

// Create takes the representation of a kongCredential and creates it.  Returns the server's representation of the kongCredential, and an error, if there is any.
func (c *FakeKongCredentials) Create(kongCredential *credential_v1.KongCredential) (result *credential_v1.KongCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kongcredentialsResource, c.ns, kongCredential), &credential_v1.KongCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*credential_v1.KongCredential), err
}

// Update takes the representation of a kongCredential and updates it. Returns the server's representation of the kongCredential, and an error, if there is any.
func (c *FakeKongCredentials) Update(kongCredential *credential_v1.KongCredential) (result *credential_v1.KongCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kongcredentialsResource, c.ns, kongCredential), &credential_v1.KongCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*credential_v1.KongCredential), err
}

// Delete takes name of the kongCredential and deletes it. Returns an error if one occurs.
func (c *FakeKongCredentials) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kongcredentialsResource, c.ns, name), &credential_v1.KongCredential{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKongCredentials) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kongcredentialsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &credential_v1.KongCredentialList{})
	return err
}

// Patch applies the patch and returns the patched kongCredential.
func (c *FakeKongCredentials) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *credential_v1.KongCredential, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kongcredentialsResource, c.ns, name, data, subresources...), &credential_v1.KongCredential{})

	if obj == nil {
		return nil, err
	}
	return obj.(*credential_v1.KongCredential), err
}
