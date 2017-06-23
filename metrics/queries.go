package metrics

import "github.com/ARGOeu/argo-messaging/stores"

func GetProjectTopics(projectUUID string, store stores.Store) (int64, error) {
	topics, err := store.QueryTopics(projectUUID, "")
	return int64(len(topics)), err
}

func GetProjectTopicsACL(projectUUID string, username string, store stores.Store) (int64, error) {
	topics, err := store.QueryTopicsByACL(projectUUID, username)
	return int64(len(topics)), err
}
