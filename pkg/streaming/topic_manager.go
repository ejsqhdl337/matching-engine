package streaming

import (
	"fmt"
	"sync"
)

type TopicManager struct {
	topics map[string]*Topic
	mutex  sync.RWMutex
}

func NewTopicManager(topics []*Topic) *TopicManager {
	topicMap := make(map[string]*Topic)
	for _, topic := range topics {
		topicMap[topic.Name] = topic
	}
	return &TopicManager{
		topics: topicMap,
	}
}

func (tm *TopicManager) GetTopic(name string) (*Topic, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	topic, ok := tm.topics[name]
	if !ok {
		return nil, fmt.Errorf("topic not found: %s", name)
	}
	return topic, nil
}
