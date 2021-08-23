package metrics

import (
	"fmt"
	amsProjects "github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	"math"
	"time"
)

func GetProjectTopics(projectUUID string, store stores.Store) (int64, error) {
	topics, _, _, err := store.QueryTopics(projectUUID, "", "", "", 0)
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
	subs, _, _, err := store.QuerySubs(projectUUID, "", "", "", 0)
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

// GetVAReport returns a VAReport populated with the needed metrics
func GetVAReport(projects []string, startDate time.Time, endDate time.Time, str stores.Store) (VAReport, error) {

	vaReport := VAReport{}

	// for the counters we need to include the ones created up to the end of the end date
	// if some gives 2020-15-01 we need to get all counters up to 2020-15-01T23:59:59
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	tpm, err := GenerateVAReport(projects, startDate, endDate, str)
	if err != nil {
		return vaReport, err
	}

	return tpm, nil
}

// GenerateVAReport returns per project metrics regarding users,topics subscriptions for the given time period
// It also includes various totals that derive from the each individual's project metrics.
// The generated result is called a VA Report
func GenerateVAReport(projects []string, startDate time.Time, endDate time.Time, str stores.Store) (VAReport, error) {

	tpj := TotalProjectsMessageCount{
		Projects:   []ProjectMetrics{},
		TotalCount: 0,
	}

	vaReport := VAReport{}

	var qtpj []stores.QProjectMessageCount
	var err error

	// since we want to present the end result using project names and not uuids
	// we need to hold the mapping of UUID to NAME
	projectsUUIDNames := make(map[string]string)

	// check that all project UUIDs are correct
	// translate the project NAMES to their respective UUIDs
	projectUUIDs := make([]string, 0)
	for _, prj := range projects {
		projectUUID := amsProjects.GetUUIDByName(prj, str)
		if projectUUID == "" {
			return VAReport{}, fmt.Errorf("Project %v", prj)
		}
		projectUUIDs = append(projectUUIDs, projectUUID)
		projectsUUIDNames[projectUUID] = prj
	}

	qtpj, err = str.QueryTotalMessagesPerProject(projectUUIDs, startDate, endDate)
	if err != nil {
		return VAReport{}, err
	}

	topicsCount, err := str.TopicsCount(startDate, endDate, projectUUIDs)
	if err != nil {
		return VAReport{}, err
	}

	subCount, err := str.SubscriptionsCount(startDate, endDate, projectUUIDs)
	if err != nil {
		return VAReport{}, err
	}

	userCount, err := str.UsersCount(startDate, endDate, projectUUIDs)
	if err != nil {
		return VAReport{}, err
	}

	for _, prj := range qtpj {

		topicCount := topicsCount[prj.ProjectUUID]
		subCount := subCount[prj.ProjectUUID]
		userCount := userCount[prj.ProjectUUID]

		projectName := ""

		// if no project names were provided we have to do the mapping between name and uuid
		if len(projects) == 0 {
			projectName = amsProjects.GetNameByUUID(prj.ProjectUUID, str)
		} else {
			projectName = projectsUUIDNames[prj.ProjectUUID]
		}

		avg := math.Ceil(prj.AverageDailyMessages*100) / 100

		pc := ProjectMessageCount{
			Project:              projectName,
			MessageCount:         prj.NumberOfMessages,
			AverageDailyMessages: avg,
		}

		projectMetrics := ProjectMetrics{
			pc,
			topicCount,
			subCount,
			userCount,
		}

		tpj.Projects = append(tpj.Projects, projectMetrics)

		tpj.TotalCount += prj.NumberOfMessages

		tpj.AverageDailyMessages += avg

		vaReport.TopicsCount += topicCount

		vaReport.SubscriptionsCount += subCount

		vaReport.UsersCount += userCount
	}

	vaReport.ProjectsMetrics = tpj

	return vaReport, nil
}
