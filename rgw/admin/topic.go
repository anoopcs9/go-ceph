//go:build ceph_preview

package admin

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
)

// Topic represents an SNS-compatible topic
type Topic struct {
	User       string        `xml:"User"`
	Name       string        `xml:"Name"`
	EndPoint   TopicEndPoint `xml:"EndPoint"`
	TopicArn   string        `xml:"TopicArn"`
	OpaqueData string        `xml:"OpaqueData"`
	Policy     string        `xml:"Policy"`
}

// TopicEndPoint represents the endpoint configuration for a topic
type TopicEndPoint struct {
	EndpointAddress    string `xml:"EndpointAddress"`
	EndpointArgs       string `xml:"EndpointArgs"`
	EndpointTopic      string `xml:"EndpointTopic"`
	HasStoredSecret    string `xml:"HasStoredSecret"`
	Persistent         string `xml:"Persistent"`
	TimeToLive         string `xml:"TimeToLive"`
	MaxRetries         string `xml:"MaxRetries"`
	RetrySleepDuration string `xml:"RetrySleepDuration"`
}

// CreateTopicResponse represents the response from a CreateTopic request
type CreateTopicResponse struct {
	XMLName           xml.Name `xml:"CreateTopicResponse"`
	CreateTopicResult struct {
		TopicArn string `xml:"TopicArn"`
	} `xml:"CreateTopicResult"`
}

// GetTopicAttributesResponse represents the response from a GetTopicAttributes request
type GetTopicAttributesResponse struct {
	XMLName                  xml.Name `xml:"GetTopicAttributesResponse"`
	GetTopicAttributesResult struct {
		Attributes struct {
			Entry []TopicAttribute `xml:"entry"`
		} `xml:"Attributes"`
	} `xml:"GetTopicAttributesResult"`
}

// TopicAttribute represents a single key-value attribute of a topic
type TopicAttribute struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

// ListTopicsResponse represents the response from a ListTopics request
type ListTopicsResponse struct {
	XMLName          xml.Name `xml:"ListTopicsResponse"`
	ListTopicsResult struct {
		Topics struct {
			Topic []Topic `xml:"member"`
		} `xml:"Topics"`
	} `xml:"ListTopicsResult"`
}

// CreateTopic creates a new SNS-compatible topic
// https://docs.ceph.com/en/latest/radosgw/s3/bucketops/#create-topic
func (api *API) CreateTopic(ctx context.Context, name string, attrs map[string]string) (string, error) {
	params := url.Values{}
	params.Set("Name", name)

	for k, v := range attrs {
		params.Set("Attributes.entry.1.key", k)
		params.Set("Attributes.entry.1.value", v)
	}

	body, err := api.callSNS(ctx, "CreateTopic", params)
	if err != nil {
		return "", err
	}

	var resp CreateTopicResponse
	err = xml.Unmarshal(body, &resp)
	if err != nil {
		return "", fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return resp.CreateTopicResult.TopicArn, nil
}

// GetTopicAttributes returns information about a specific topic
func (api *API) GetTopicAttributes(ctx context.Context, topicARN string) (Topic, error) {
	params := url.Values{}
	params.Set("TopicArn", topicARN)

	body, err := api.callSNS(ctx, "GetTopicAttributes", params)
	if err != nil {
		return Topic{}, err
	}

	var resp GetTopicAttributesResponse
	err = xml.Unmarshal(body, &resp)
	if err != nil {
		return Topic{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	var topic Topic
	for _, entry := range resp.GetTopicAttributesResult.Attributes.Entry {
		switch entry.Key {
		case "User":
			topic.User = entry.Value
		case "Name":
			topic.Name = entry.Value
		case "TopicArn":
			topic.TopicArn = entry.Value
		case "OpaqueData":
			topic.OpaqueData = entry.Value
		case "Policy":
			topic.Policy = entry.Value
		}
	}

	return topic, nil
}

// DeleteTopic deletes a topic
func (api *API) DeleteTopic(ctx context.Context, topicARN string) error {
	params := url.Values{}
	params.Set("TopicArn", topicARN)

	_, err := api.callSNS(ctx, "DeleteTopic", params)
	return err
}

// ListTopics lists all topics for the tenant
func (api *API) ListTopics(ctx context.Context) ([]Topic, error) {
	body, err := api.callSNS(ctx, "ListTopics", nil)
	if err != nil {
		return nil, err
	}

	var resp ListTopicsResponse
	err = xml.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return resp.ListTopicsResult.Topics.Topic, nil
}
