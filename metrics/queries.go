package metrics

import "github.com/ARGOeu/argo-messaging/stores"

func GetProjectTopics(projectUUID string, store stores.Store) (int64, error) {
	topics, err := store.QueryTopics(projectUUID, "")
	return int64(len(topics)), err
}

func GetProjectSubsByTopic(projectUUID string, topic string, store stores.Store) (int64, error) {
	subs, err := store.QuerySubsByTopic(projectUUID, topic)
	return int64(len(subs)), err
}

func GetProjectTopicsACL(projectUUID string, username string, store stores.Store) (int64, error) {
	topics, err := store.QueryTopicsByACL(projectUUID, username)
	return int64(len(topics)), err
}

func GetProjectSubs(projectUUID string, store stores.Store) (int64, error) {
	subs, err := store.QuerySubs(projectUUID, "")
	return int64(len(subs)), err
}

func GetProjectSubsACL(projectUUID string, username string, store stores.Store) (int64, error) {
	subs, err := store.QuerySubsByACL(projectUUID, username)
	return int64(len(subs)), err
}
