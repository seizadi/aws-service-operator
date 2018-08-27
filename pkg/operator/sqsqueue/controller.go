// >>>>>>> DO NOT EDIT THIS FILE <<<<<<<<<<
// This file is autogenerated via `aws-operator generate`
// If you'd like the change anything about this file make edits to the .templ
// file in the pkg/codegen/assets directory.

package sqsqueue

import (
	 metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/christopherhein/aws-operator/pkg/helpers"
	"reflect"

	"github.com/christopherhein/aws-operator/pkg/config"
  "github.com/christopherhein/aws-operator/pkg/queue"
	opkit "github.com/christopherhein/operator-kit"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/tools/cache"

	awsapi "github.com/christopherhein/aws-operator/pkg/apis/operator.aws"
	awsV1alpha1 "github.com/christopherhein/aws-operator/pkg/apis/operator.aws/v1alpha1"
	awsclient "github.com/christopherhein/aws-operator/pkg/client/clientset/versioned/typed/operator.aws/v1alpha1"
)

// Resource is the object store definition
var Resource = opkit.CustomResource{
	Name:       "sqsqueue",
	Plural:     "sqsqueues",
	Group:      awsapi.GroupName,
	Version:    awsapi.Version,
	Scope:      apiextensionsv1beta1.NamespaceScoped,
	Kind:       reflect.TypeOf(awsV1alpha1.SQSQueue{}).Name(),
	ShortNames: []string{
		"sqs",
		"queue",
		"queues",
	},
}

// Controller represents a controller object for object store custom resources
type Controller struct {
	config       *config.Config
	context      *opkit.Context
	awsclientset awsclient.OperatorV1alpha1Interface
  topicARN     string
}

// NewController create controller for watching object store custom resources created
func NewController(config *config.Config, context *opkit.Context, awsclientset awsclient.OperatorV1alpha1Interface) *Controller {
	return &Controller{
		config:       config,
		context:      context,
		awsclientset: awsclientset,
	}
}

// StartWatch watches for instances of Object Store custom resources and acts on them
func (c *Controller) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}
	queuectrl := queue.New(c.config, c.context, c.awsclientset, 1)
	c.topicARN, _, _, _ = queuectrl.Register("sqsqueue", &awsV1alpha1.SQSQueue{})
	go queuectrl.StartWatch(queue.HandlerFunc(QueueUpdater), stopCh)

	restClient := c.awsclientset.RESTClient()
	watcher := opkit.NewWatcher(Resource, namespace, resourceHandlers, restClient)
	go watcher.Watch(&awsV1alpha1.SQSQueue{}, stopCh)

	return nil
}
// QueueUpdater will take the messages from the queue and process them
func QueueUpdater(config *config.Config, msg *queue.MessageBody) error {
	logger := config.Logger
	var name, namespace string
	if msg.Updatable {
		name = msg.ResourceName
		namespace = msg.Namespace
	} else {
		clientSet, _ := awsclient.NewForConfig(config.RESTConfig)
		resources, err := clientSet.SQSQueues("").List(metav1.ListOptions{})
		if err != nil {
			logger.WithError(err).Error("error getting sqsqueues")
			return err
		}
		for _, resource := range resources.Items {
			if resource.Status.StackID == msg.ParsedMessage["StackId"] {
				name = resource.Name
				namespace = resource.Namespace
			}
		}
	}

	if name != "" && namespace != "" {
		if msg.ParsedMessage["ResourceStatus"] == "ROLLBACK_COMPLETE" {
			err := deleteStack(config, name, namespace, msg.ParsedMessage["StackId"])
			if err != nil {
				return err
			}
		} else if msg.ParsedMessage["ResourceStatus"] == "DELETE_COMPLETE" {
			err := updateStatus(config, name, namespace, msg.ParsedMessage["StackId"], msg.ParsedMessage["ResourceStatus"], msg.ParsedMessage["ResourceStatusReason"])
			if err != nil {
				return err
			}

			err = incrementRollbackCount(config, name, namespace)
			if err != nil {
				return err
			}
		} else {
			err := updateStatus(config, name, namespace, msg.ParsedMessage["StackId"], msg.ParsedMessage["ResourceStatus"], msg.ParsedMessage["ResourceStatusReason"])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Controller) onAdd(obj interface{}) {
	s := obj.(*awsV1alpha1.SQSQueue).DeepCopy()
  if s.Status.ResourceStatus == "" || s.Status.ResourceStatus == "DELETE_COMPLETE" {
    cft := New(c.config, s, c.topicARN)
    output, err := cft.CreateStack()
    if err != nil {
      c.config.Logger.WithError(err).Errorf("error creating sqsqueue '%s'", s.Name)
      return
    }
    c.config.Logger.Infof("added sqsqueue '%s' with stackID '%s'", s.Name, string(*output.StackId))
    c.config.Logger.Infof("view at https://console.aws.amazon.com/cloudformation/home?#/stack/detail?stackId=%s", string(*output.StackId))

		err = updateStatus(c.config, s.Name, s.Namespace, string(*output.StackId), "CREATE_IN_PROGRESS", "")
		if err != nil {
			c.config.Logger.WithError(err).Error("error updating status")
		}
  }
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	oo := oldObj.(*awsV1alpha1.SQSQueue).DeepCopy()
	no := newObj.(*awsV1alpha1.SQSQueue).DeepCopy()

	if no.Status.ResourceStatus == "DELETE_COMPLETE" {
		c.onAdd(no)
  }
  if helpers.IsStackComplete(oo.Status.ResourceStatus, false) && !reflect.DeepEqual(oo.Spec, no.Spec) {
    cft := New(c.config, oo, c.topicARN)
    output, err := cft.UpdateStack(no)
    if err != nil {
      c.config.Logger.WithError(err).Errorf("error updating sqsqueue '%s' with new params %+v and old %+v", no.Name, no, oo)
      return
    }
    c.config.Logger.Infof("updated sqsqueue '%s' with params '%s'", no.Name, string(*output.StackId))
    c.config.Logger.Infof("view at https://console.aws.amazon.com/cloudformation/home?#/stack/detail?stackId=%s", string(*output.StackId))

		err = updateStatus(c.config, oo.Name, oo.Namespace, string(*output.StackId), "UPDATE_IN_PROGRESS", "")
		if err != nil {
			c.config.Logger.WithError(err).Error("error updating status")
		}
  }
}

func (c *Controller) onDelete(obj interface{}) {
	s := obj.(*awsV1alpha1.SQSQueue).DeepCopy()
	cft := New(c.config, s, c.topicARN)
	err := cft.DeleteStack()
	if err != nil {
		c.config.Logger.WithError(err).Errorf("error deleting sqsqueue '%s'", s.Name)
		return
	}

	c.config.Logger.Infof("deleted sqsqueue '%s'", s.Name)
}
func incrementRollbackCount(config *config.Config, name string, namespace string) error {
	logger := config.Logger
	clientSet, _ := awsclient.NewForConfig(config.RESTConfig)
	resource, err := clientSet.SQSQueues(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		logger.WithError(err).Error("error getting sqsqueues")
		return err
	}

	resourceCopy := resource.DeepCopy()
	resourceCopy.Spec.RollbackCount = resourceCopy.Spec.RollbackCount+1

	_, err = clientSet.SQSQueues(namespace).Update(resourceCopy)
	if err != nil {
		logger.WithError(err).Error("error updating resource")
		return err
	}
	return nil
}

func updateStatus(config *config.Config, name string, namespace string, stackID string, status string, reason string) error {
	logger := config.Logger
	clientSet, _ := awsclient.NewForConfig(config.RESTConfig)
	resource, err := clientSet.SQSQueues(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		logger.WithError(err).Error("error getting sqsqueues")
		return err
	}

	resourceCopy := resource.DeepCopy()
	resourceCopy.Status.ResourceStatus = status
	resourceCopy.Status.ResourceStatusReason = reason
	resourceCopy.Status.StackID = stackID

	if helpers.IsStackComplete(status, false) {
		cft := New(config, resourceCopy, "")
		outputs, err := cft.GetOutputs()
		if err != nil {
			logger.WithError(err).Error("error getting outputs")
		}
		resourceCopy.Output.QueueURL = outputs["QueueURL"]
		resourceCopy.Output.QueueARN = outputs["QueueARN"]
		resourceCopy.Output.QueueName = outputs["QueueName"]
		resourceCopy.Output.DeadLetterQueueURL = outputs["DeadLetterQueueURL"]
		resourceCopy.Output.DeadLetterQueueARN = outputs["DeadLetterQueueARN"]
		resourceCopy.Output.DeadLetterQueueName = outputs["DeadLetterQueueName"]
	}

	_, err = clientSet.SQSQueues(namespace).Update(resourceCopy)
	if err != nil {
		logger.WithError(err).Error("error updating resource")
		return err
	}

	err = syncAdditionalResources(config, resourceCopy)
	if err != nil {
		logger.WithError(err).Info("error syncing resources")
	}
	return nil
}

func deleteStack(config *config.Config, name string, namespace string, stackID string) error {
	logger := config.Logger
	clientSet, _ := awsclient.NewForConfig(config.RESTConfig)
	resource, err := clientSet.SQSQueues(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		logger.WithError(err).Error("error getting sqsqueues")
		return err
	}

	cft := New(config, resource, "")
	err = cft.DeleteStack()
	if err != nil {
		return err
	}

	err = cft.WaitUntilStackDeleted()
	return err
}

func syncAdditionalResources(config *config.Config, s *awsV1alpha1.SQSQueue) (err error) {
	clientSet, _ := awsclient.NewForConfig(config.RESTConfig)
	resource, err := clientSet.SQSQueues(s.Namespace).Get(s.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	resource = resource.DeepCopy()

	
	


	_, err = clientSet.SQSQueues(s.Namespace).Update(resource)
	if err != nil {
		return err
	}
  return nil
}
