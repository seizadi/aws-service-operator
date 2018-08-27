// >>>>>>> DO NOT EDIT THIS FILE <<<<<<<<<<
// This file is autogenerated via `aws-operator-codegen process`
// If you'd like the change anything about this file make edits to the .templ
// file in the pkg/codegen/assets directory.

package snssubscription

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	awsV1alpha1 "github.com/christopherhein/aws-operator/pkg/apis/operator.aws/v1alpha1"
	"github.com/christopherhein/aws-operator/pkg/config"
	"github.com/christopherhein/aws-operator/pkg/helpers"
)

// New generates a new object
func New(config *config.Config, snssubscription *awsV1alpha1.SNSSubscription, topicARN string) *Cloudformation {
	return &Cloudformation{
		SNSSubscription: snssubscription,
		config:					config,
    topicARN:       topicARN,
	}
}

// Cloudformation defines the snssubscription cfts
type Cloudformation struct {
	config         *config.Config
	SNSSubscription *awsV1alpha1.SNSSubscription
  topicARN       string
}

// StackName returns the name of the stack based on the aws-operator-config
func (s *Cloudformation) StackName() string {
	return helpers.StackName(s.config.ClusterName, "snssubscription", s.SNSSubscription.Name, s.SNSSubscription.Namespace)
}

// GetOutputs return the stack outputs from the DescribeStacks call
func (s *Cloudformation) GetOutputs() (map[string]string, error) {
	outputs := map[string]string{}
	sess := s.config.AWSSession
	svc := cloudformation.New(sess)

	stackInputs := cloudformation.DescribeStacksInput{
		StackName:   aws.String(s.StackName()),
	}

	output, err := svc.DescribeStacks(&stackInputs)
	if err != nil {
		return nil, err
	}
	// Not sure if this is even possible
	if len(output.Stacks) != 1 {
		return nil, errors.New("no stacks returned with that stack name")
	}

	for _, out := range output.Stacks[0].Outputs {
		outputs[*out.OutputKey] = *out.OutputValue
	}

	return outputs, err
}

// CreateStack will create the stack with the supplied params
func (s *Cloudformation) CreateStack() (output *cloudformation.CreateStackOutput, err error) {
	sess := s.config.AWSSession
	svc := cloudformation.New(sess)

	cftemplate := helpers.GetCloudFormationTemplate(s.config, "snssubscription", s.SNSSubscription.Spec.CloudFormationTemplateName, s.SNSSubscription.Spec.CloudFormationTemplateNamespace)

	stackInputs := cloudformation.CreateStackInput{
		StackName:   aws.String(s.StackName()),
		TemplateURL: aws.String(cftemplate),
		NotificationARNs: []*string{
			aws.String(s.topicARN),
		},
	}

	resourceName := helpers.CreateParam("ResourceName", s.SNSSubscription.Name)
	resourceVersion := helpers.CreateParam("ResourceVersion", s.SNSSubscription.ResourceVersion)
	namespace := helpers.CreateParam("Namespace", s.SNSSubscription.Namespace)
	clusterName := helpers.CreateParam("ClusterName", s.config.ClusterName)
	topicNameTemp :=	"{{(call .Helpers.GetSNSTopicByName .Config .Obj.Spec.TopicName .Obj.Namespace).Output.TopicARN}}"
	topicNameValue, err := helpers.Templatize(topicNameTemp, helpers.Data{Obj: s.SNSSubscription, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	topicName := helpers.CreateParam("TopicARN", helpers.Stringify(topicNameValue))
	protocolTemp :=	"{{.Obj.Spec.Protocol}}"
	protocolValue, err := helpers.Templatize(protocolTemp, helpers.Data{Obj: s.SNSSubscription, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	protocol := helpers.CreateParam("Protocol", helpers.Stringify(protocolValue))
	endpointTemp :=	"{{if (eq .Obj.Spec.Protocol \"sqs\")}}{{(call .Helpers.GetSQSQueueByName .Config .Obj.Spec.Endpoint .Obj.Namespace).Output.QueueARN }}{{else}}{{.Obj.Spec.Endpoint}}{{end}}"
	endpointValue, err := helpers.Templatize(endpointTemp, helpers.Data{Obj: s.SNSSubscription, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	endpoint := helpers.CreateParam("Endpoint", helpers.Stringify(endpointValue))
	queueURLTemp :=	"{{(call .Helpers.GetSQSQueueByName .Config .Obj.Spec.Endpoint .Obj.Namespace).Output.QueueURL }}"
	queueURLValue, err := helpers.Templatize(queueURLTemp, helpers.Data{Obj: s.SNSSubscription, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	queueURL := helpers.CreateParam("QueueURL", helpers.Stringify(queueURLValue))

	parameters := []*cloudformation.Parameter{}
	parameters = append(parameters, resourceName)
	parameters = append(parameters, resourceVersion)
	parameters = append(parameters, namespace)
	parameters = append(parameters, clusterName)
	parameters = append(parameters, topicName)
	parameters = append(parameters, protocol)
	parameters = append(parameters, endpoint)
	parameters = append(parameters, queueURL)

	stackInputs.SetParameters(parameters)

	resourceNameTag := helpers.CreateTag("ResourceName", s.SNSSubscription.Name)
	resourceVersionTag := helpers.CreateTag("ResourceVersion", s.SNSSubscription.ResourceVersion)
	namespaceTag := helpers.CreateTag("Namespace", s.SNSSubscription.Namespace)
	clusterNameTag := helpers.CreateTag("ClusterName", s.config.ClusterName)

	tags := []*cloudformation.Tag{}
	tags = append(tags, resourceNameTag)
	tags = append(tags, resourceVersionTag)
	tags = append(tags, namespaceTag)
	tags = append(tags, clusterNameTag)

	stackInputs.SetTags(tags)

  output, err = svc.CreateStack(&stackInputs)
	return
}

// UpdateStack will update the existing stack
func (s *Cloudformation) UpdateStack(updated *awsV1alpha1.SNSSubscription) (output *cloudformation.UpdateStackOutput, err error) {
	sess := s.config.AWSSession
	svc := cloudformation.New(sess)

	cftemplate := helpers.GetCloudFormationTemplate(s.config, "snssubscription", updated.Spec.CloudFormationTemplateName, updated.Spec.CloudFormationTemplateNamespace)

	stackInputs := cloudformation.UpdateStackInput{
		StackName:   aws.String(s.StackName()),
		TemplateURL: aws.String(cftemplate),
		NotificationARNs: []*string{
			aws.String(s.topicARN),
		},
	}

	resourceName := helpers.CreateParam("ResourceName", s.SNSSubscription.Name)
	resourceVersion := helpers.CreateParam("ResourceVersion", s.SNSSubscription.ResourceVersion)
	namespace := helpers.CreateParam("Namespace", s.SNSSubscription.Namespace)
	clusterName := helpers.CreateParam("ClusterName", s.config.ClusterName)
	topicNameTemp := "{{(call .Helpers.GetSNSTopicByName .Config .Obj.Spec.TopicName .Obj.Namespace).Output.TopicARN}}"
	topicNameValue, err := helpers.Templatize(topicNameTemp, helpers.Data{Obj: updated, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	topicName := helpers.CreateParam("TopicARN", helpers.Stringify(topicNameValue))
	protocolTemp :=	"{{.Obj.Spec.Protocol}}"
	protocolValue, err := helpers.Templatize(protocolTemp, helpers.Data{Obj: updated, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	protocol := helpers.CreateParam("Protocol", helpers.Stringify(protocolValue))
	endpointTemp := "{{if (eq .Obj.Spec.Protocol \"sqs\")}}{{(call .Helpers.GetSQSQueueByName .Config .Obj.Spec.Endpoint .Obj.Namespace).Output.QueueARN }}{{else}}{{.Obj.Spec.Endpoint}}{{end}}"
	endpointValue, err := helpers.Templatize(endpointTemp, helpers.Data{Obj: updated, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	endpoint := helpers.CreateParam("Endpoint", helpers.Stringify(endpointValue))
	queueURLTemp := "{{(call .Helpers.GetSQSQueueByName .Config .Obj.Spec.Endpoint .Obj.Namespace).Output.QueueURL }}"
	queueURLValue, err := helpers.Templatize(queueURLTemp, helpers.Data{Obj: updated, Config: s.config, Helpers: helpers.New()})
	if err != nil {
		return output, err
	}
	queueURL := helpers.CreateParam("QueueURL", helpers.Stringify(queueURLValue))

	parameters := []*cloudformation.Parameter{}
	parameters = append(parameters, resourceName)
	parameters = append(parameters, resourceVersion)
	parameters = append(parameters, namespace)
	parameters = append(parameters, clusterName)
	parameters = append(parameters, topicName)
	parameters = append(parameters, protocol)
	parameters = append(parameters, endpoint)
	parameters = append(parameters, queueURL)

	stackInputs.SetParameters(parameters)

	resourceNameTag := helpers.CreateTag("ResourceName", s.SNSSubscription.Name)
	resourceVersionTag := helpers.CreateTag("ResourceVersion", s.SNSSubscription.ResourceVersion)
	namespaceTag := helpers.CreateTag("Namespace", s.SNSSubscription.Namespace)
	clusterNameTag := helpers.CreateTag("ClusterName", s.config.ClusterName)

	tags := []*cloudformation.Tag{}
	tags = append(tags, resourceNameTag)
	tags = append(tags, resourceVersionTag)
	tags = append(tags, namespaceTag)
	tags = append(tags, clusterNameTag)

	stackInputs.SetTags(tags)

  output, err = svc.UpdateStack(&stackInputs)
	return
}

// DeleteStack will delete the stack
func (s *Cloudformation) DeleteStack() (err error) {
	sess := s.config.AWSSession
	svc := cloudformation.New(sess)

	stackInputs := cloudformation.DeleteStackInput{}
	stackInputs.SetStackName(s.StackName())

  _, err = svc.DeleteStack(&stackInputs)
	return
}

// WaitUntilStackDeleted will delete the stack
func (s *Cloudformation) WaitUntilStackDeleted() (err error) {
	sess := s.config.AWSSession
	svc := cloudformation.New(sess)

	stackInputs := cloudformation.DescribeStacksInput{
		StackName:   aws.String(s.StackName()),
	}

  err = svc.WaitUntilStackDeleteComplete(&stackInputs)
	return
}
