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
	consumer_v1 "github.com/kong/ingress-controller/internal/apis/consumer/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKongConsumers implements KongConsumerInterface
type FakeKongConsumers struct {
	Fake *FakeConfigurationV1
	ns   string
}

var kongconsumersResource = schema.GroupVersionResource{Group: "configuration.konghq.com", Version: "v1", Resource: "kongconsumers"}

var kongconsumersKind = schema.GroupVersionKind{Group: "configuration.konghq.com", Version: "v1", Kind: "KongConsumer"}

// Get takes name of the kongConsumer, and returns the corresponding kongConsumer object, and an error if there is any.
func (c *FakeKongConsumers) Get(name string, options v1.GetOptions) (result *consumer_v1.KongConsumer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kongconsumersResource, c.ns, name), &consumer_v1.KongConsumer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*consumer_v1.KongConsumer), err
}

// List takes label and field selectors, and returns the list of KongConsumers that match those selectors.
func (c *FakeKongConsumers) List(opts v1.ListOptions) (result *consumer_v1.KongConsumerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kongconsumersResource, kongconsumersKind, c.ns, opts), &consumer_v1.KongConsumerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &consumer_v1.KongConsumerList{}
	for _, item := range obj.(*consumer_v1.KongConsumerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kongConsumers.
func (c *FakeKongConsumers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kongconsumersResource, c.ns, opts))

}

// Create takes the representation of a kongConsumer and creates it.  Returns the server's representation of the kongConsumer, and an error, if there is any.
func (c *FakeKongConsumers) Create(kongConsumer *consumer_v1.KongConsumer) (result *consumer_v1.KongConsumer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kongconsumersResource, c.ns, kongConsumer), &consumer_v1.KongConsumer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*consumer_v1.KongConsumer), err
}

// Update takes the representation of a kongConsumer and updates it. Returns the server's representation of the kongConsumer, and an error, if there is any.
func (c *FakeKongConsumers) Update(kongConsumer *consumer_v1.KongConsumer) (result *consumer_v1.KongConsumer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kongconsumersResource, c.ns, kongConsumer), &consumer_v1.KongConsumer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*consumer_v1.KongConsumer), err
}

// Delete takes name of the kongConsumer and deletes it. Returns an error if one occurs.
func (c *FakeKongConsumers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kongconsumersResource, c.ns, name), &consumer_v1.KongConsumer{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKongConsumers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kongconsumersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &consumer_v1.KongConsumerList{})
	return err
}

// Patch applies the patch and returns the patched kongConsumer.
func (c *FakeKongConsumers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *consumer_v1.KongConsumer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kongconsumersResource, c.ns, name, data, subresources...), &consumer_v1.KongConsumer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*consumer_v1.KongConsumer), err
}
