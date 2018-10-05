package metrics

import (
	"github.com/ARGOeu/argo-messaging/stores"
	"time"
)

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

func GetDailyTopicMsgCount(projectUUID string, topicName string, store stores.Store) ([]Timepoint, error) {

	var err error
	var qDtmc []stores.QDailyTopicMsgCount

	timePoints := []Timepoint{}

	if qDtmc, err = store.QueryDailyTopicMsgCount(projectUUID, topicName, time.Time{}); err != nil {
		return timePoints, err
	}
	for _, qd := range qDtmc {
		timePoints = append(timePoints, Timepoint{qd.Date.Format("2006-01-02"), qd.NumberOfMessages})
	}

	return timePoints, err
}

func GetDailyProjectMsgCount(projectUUID string, store stores.Store) ([]Timepoint, error) {

	var err error
	var qDpmc []stores.QDailyProjectMsgCount

	timePoints := []Timepoint{}

	if qDpmc, err = store.QueryDailyProjectMsgCount(projectUUID); err != nil {
		return timePoints, err
	}

	for _, qdp := range qDpmc {
		timePoints = append(timePoints, Timepoint{qdp.Date.Format("2006-01-02"), qdp.NumberOfMessages})
	}

	return timePoints, err
}

func AggrProjectUserSubs(projectUUID string, store stores.Store) (MetricList, error) {
	pr, err := store.QueryProjects(projectUUID, "")
	if err != nil {
		return MetricList{}, err
	}
	prName := pr[0].Name
	users, err := store.QueryUsers(projectUUID, "", "")
	ml := MetricList{}
	for _, item := range users {
		username := item.Name
		userUUID := item.UUID
		numSubs, _ := GetProjectSubsACL(projectUUID, userUUID, store)
		if numSubs > 0 {
			m := NewProjectUserSubs(prName, username, numSubs, GetTimeNowZulu())
			ml.Metrics = append(ml.Metrics, m)
		}

	}
	return ml, err
}

func AggrProjectUserTopics(projectUUID string, store stores.Store) (MetricList, error) {
	pr, err := store.QueryProjects(projectUUID, "")
	if err != nil {
		return MetricList{}, err
	}
	prName := pr[0].Name
	users, err := store.QueryUsers(projectUUID, "", "")
	ml := MetricList{}
	for _, item := range users {
		username := item.Name
		userUUID := item.UUID
		numSubs, _ := GetProjectTopicsACL(projectUUID, userUUID, store)
		if numSubs > 0 {
			m := NewProjectUserTopics(prName, username, numSubs, GetTimeNowZulu())
			ml.Metrics = append(ml.Metrics, m)
		}

	}
	return ml, err
}
