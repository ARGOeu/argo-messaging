package push

import (
	"errors"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/messages"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
)

// Pusher holds information for the pusher routine and subscription
type Pusher struct {
	id          int
	sub         subscriptions.Subscription
	endpoint    string
	retryPolicy string
	retryPeriod int
	stop        chan int      // 1: Stop 2: restart
	rate        time.Duration // in milliseconds
	running     bool
	sndr        Sender
	mgr         *Manager
}

// Manager manages all pusher routines
type Manager struct {
	list   map[string]*Pusher // map using as key the string = "{project}/{sub}"
	broker brokers.Broker     // Reference to backend broker
	store  stores.Store       // Reference to backend store
	sender Sender             // Reference to send mechanism (HTTP client)
}

// LoadPushSubs is called during API initialization to retrieve available
// push configured subs and activate them
func (mgr *Manager) LoadPushSubs() {

	results := subscriptions.LoadPushSubs(mgr.store)

	// Add all of them
	for _, item := range results.List {
		mgr.Add(item.ProjectUUID, item.Name)
	}
}

// StartAll enables all pushsers
func (mgr *Manager) StartAll() {
	for k := range mgr.list {
		item := mgr.list[k]
		item.launch(mgr.broker, mgr.store.Clone())
	}
}

// StopAll stops Activity on all pushers
func (mgr *Manager) StopAll() error {
	for k := range mgr.list {
		project, sub, err := splitPSub(k)
		if err != nil {
			return err
		}
		mgr.Stop(project, sub)
	}
	return nil
}

// RemoveProjectAll stops and removes all pushers related to a project
func (mgr *Manager) RemoveProjectAll(projectUUID string) error {
	// collect all subs to be removed here
	subsToRemove := []string{}
	// Iterate and stop all relevant subs
	for k := range mgr.list {
		project, sub, err := splitPSub(k)
		if err != nil {
			return err
		}
		if project == projectUUID {
			mgr.Stop(projectUUID, sub)
			subsToRemove = append(subsToRemove, sub)
		}

	}
	// Now remove relevant subs from the list
	for _, sub := range subsToRemove {
		mgr.Remove(projectUUID, sub)
	}

	return nil
}

// Push method of pusher object to consume and push messages
func (p *Pusher) push(brk brokers.Broker, store stores.Store) {
	log.Debug("pid ", p.id, "pushing")
	// update sub details

	subs, err := subscriptions.Find(p.sub.ProjectUUID, p.sub.Name, store)

	// If subscription doesn't exist in store stop and remove it from manager
	if err == nil && len(subs.List) == 0 {
		p.stop <- 1
		return
	}
	p.sub = subs.List[0]
	// Init Received Message List

	fullTopic := p.sub.ProjectUUID + "." + p.sub.Topic
	msgs, err := brk.Consume(fullTopic, p.sub.Offset, true, 1)
	if err != nil {
		// If tracked offset is off, update it to the latest min offset
		if err == brokers.ErrOffsetOff {
			// Get Current Min Offset and advanced tracked one
			p.sub.Offset = brk.GetMinOffset(fullTopic)
			msgs, err = brk.Consume(fullTopic, p.sub.Offset, true, 1)
			if err != nil {
				log.Error("Unable to consume after updating offset")
				return
			}
		}
	}
	if len(msgs) > 0 {
		// Generate push message template
		pMsg := messages.PushMsg{}

		pMsg.Msg, _ = messages.LoadMsgJSON([]byte(msgs[0]))
		pMsg.Sub = p.sub.FullName
		pMsgJSON, _ := pMsg.ExportJSON()
		err := p.sndr.Send(pMsgJSON, p.endpoint)

		if err == nil {
			// Advance the offset
			store.UpdateSubOffset(p.sub.ProjectUUID, p.sub.Name, 1+p.sub.Offset)
			log.Debug("offset updated")
		}
	} else {
		log.Debug("pid: ", p.id, " empty")
	}
}

// PrintAll prints manager stats
func (mgr *Manager) PrintAll() {
	for k := range mgr.list {
		item := mgr.list[k]
		log.Debug("--- pid: ", item.id, " running: ", item.running)
	}
}

// NewManager creates a new manager object for managing push routines
func NewManager(brk brokers.Broker, str stores.Store, sndr Sender) *Manager {
	mgr := Manager{}
	mgr.broker = brk
	mgr.store = str
	mgr.sender = sndr
	mgr.list = make(map[string]*Pusher)
	log.Info("PUSH", "\t", "Manager Initialized")
	return &mgr
}

func splitPSub(psub string) (string, string, error) {
	tokens := strings.Split(psub, "/")
	if len(tokens) != 2 {
		return "", "", errors.New("Wrong project/subscription definition")
	}

	return tokens[0], tokens[1], nil
}

// isSet returns true if broker and store has been set
func (mgr *Manager) isSet() bool {
	if mgr.broker != nil && mgr.store != nil {
		return true
	}

	return false
}

// Get returns a pusher
func (mgr *Manager) Get(psub string) (*Pusher, error) {
	if p, ok := mgr.list[psub]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}

// Remove a push subscription
func (mgr *Manager) Remove(project string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if _, err := mgr.Get(project + "/" + sub); err == nil {
		delete(mgr.list, project+"/"+sub)
		return nil
	}

	return errors.New("not Found")
}

// Restart a push subscription
func (mgr *Manager) Restart(project string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(project + "/" + sub); err == nil {
		if p.running == false {
			log.Debug("Already stopped", p.id, "state:", p.running)
			return errors.New("Already Stoped")
		}
		log.Debug("Trying to Restart:", p.id)
		p.stop <- 2
		return nil
	}

	return errors.New("not Found")
}

// Stop stops a push subscription
func (mgr *Manager) Stop(project string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(project + "/" + sub); err == nil {
		if p.running == false {
			log.Debug("Already stopped", p.id, " state:", p.running)
			return errors.New("Already Stoped")
		}
		log.Debug("Trying to stop:", p.id)
		p.stop <- 1
		return nil
	}

	return errors.New("not Found")
}

// Refresh updates the subscription information from the database
func (mgr *Manager) Refresh(projectUUID string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(projectUUID + "/" + sub); err == nil {

		subs, err := subscriptions.Find(projectUUID, sub, mgr.store)

		if err != nil {
			return errors.New("backend error")
		}

		if subs.Empty() {
			return errors.New("Not Found")
		}

		p.endpoint = subs.List[0].PushCfg.Pend
		p.retryPolicy = subs.List[0].PushCfg.RetPol.PolicyType
		p.retryPeriod = subs.List[0].PushCfg.RetPol.Period
		p.rate = time.Duration(p.retryPeriod) * time.Millisecond
	}

	return errors.New("not Found")
}

// Add a new push subscription
func (mgr *Manager) Add(projectUUID string, subName string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}
	// Check if subscription exists
	subs, err := subscriptions.Find(projectUUID, subName, mgr.store)

	if err != nil {
		return errors.New("Backend error")
	}

	if subs.Empty() {
		return errors.New("not found")
	}

	// Create new pusher
	pushr := Pusher{}
	pushr.id = len(mgr.list)
	pushr.sub = subs.List[0]
	pushr.endpoint = subs.List[0].PushCfg.Pend
	pushr.running = false
	pushr.stop = make(chan int, 2)
	pushr.retryPolicy = subs.List[0].PushCfg.RetPol.PolicyType
	pushr.retryPeriod = subs.List[0].PushCfg.RetPol.Period
	pushr.rate = time.Duration(pushr.retryPeriod) * time.Millisecond
	pushr.sndr = mgr.sender
	pushr.mgr = mgr
	mgr.list[projectUUID+"/"+subName] = &pushr
	log.Info("PUSH", "\t", "Push Subscription Added")

	return nil

}

// Launch Launches a new puhser
func (mgr *Manager) Launch(project string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	mgr.Refresh(project, sub)

	psub := project + "/" + sub

	if p, err := mgr.Get(psub); err == nil {
		if p.running == true {
			return errors.New("Already Running")
		}

		p.launch(mgr.broker, mgr.store.Clone())
		return nil
	}

	return errors.New("not Found")
}

// Launch the pusher activity
func (p *Pusher) launch(brk brokers.Broker, store stores.Store) {
	log.Info("PUSH", "\t", "pusher: ", p.id, " launching...")
	p.running = true
	if p.retryPolicy == "linear" {
		go LinearActivity(p, brk, store)
	}

}

//LinearActivity implements a linear retry push
func LinearActivity(p *Pusher, brk brokers.Broker, store stores.Store) error {

	defer store.Close()

	for {
		rate := time.After(p.rate)
		select {
		case halt := <-p.stop:
			{

				log.Info("PUSH", "\t", "pusher: ", p.id, " stoping...")
				p.running = false
				if halt == 2 {
					p.mgr.Launch(p.sub.ProjectUUID, p.sub.Name)
				} else {
					p.mgr.Remove(p.sub.ProjectUUID, p.sub.Name)
				}
				return nil
			}
		case <-rate:
			{
				p.push(brk, store)
			}
		}
	}

}
