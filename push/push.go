package push

import (
	"errors"
	"log"
	"strings"
	"time"

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
	// Retrieve available push subscriptions
	subs := subscriptions.Subscriptions{}
	subs.LoadPushSubs(mgr.store)

	// Add all of them
	for _, item := range subs.List {
		mgr.Add(item.Project, item.Name)
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

// Push method of pusher object to consume and push messages
func (p *Pusher) push(brk brokers.Broker, store stores.Store) {
	log.Println("pid", p.id, "pushing")
	// update sub details
	subs := subscriptions.Subscriptions{}
	subs.LoadOne(p.sub.Project, p.sub.Name, store)
	p.sub = subs.List[0]
	// Init Received Message List

	fullTopic := p.sub.Project + "." + p.sub.Topic
	msgs := brk.Consume(fullTopic, p.sub.Offset, false)
	if len(msgs) > 0 {
		// Generate push message template
		pMsg := messages.PushMsg{}

		pMsg.Msg, _ = messages.LoadMsgJSON([]byte(msgs[0]))
		pMsg.Sub = p.sub.FullName
		pMsgJSON, _ := pMsg.ExportJSON()
		err := p.sndr.Send(pMsgJSON, p.endpoint)

		if err == nil {
			// Advance the offset
			store.UpdateSubOffset(p.sub.Name, 1+p.sub.Offset)
			log.Println("offset updated")
		}
	} else {
		log.Println("pid:", p.id, "empty")
	}
}

// PrintAll prints manager stats
func (mgr *Manager) PrintAll() {
	for k := range mgr.list {
		item := mgr.list[k]
		log.Println("--- pid:", item.id, "running:", item.running)
	}
}

// NewManager creates a new manager object for managing push routines
func NewManager(brk brokers.Broker, str stores.Store, sndr Sender) *Manager {
	mgr := Manager{}
	mgr.broker = brk
	mgr.store = str
	mgr.sender = sndr
	mgr.list = make(map[string]*Pusher)
	log.Println("Manager Initialized")
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

// Restart a push subscription
func (mgr *Manager) Restart(project string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(project + "/" + sub); err == nil {
		if p.running == false {
			log.Println("Already stopped", p.id, "state:", p.running)
			return errors.New("Already Stoped")
		}
		log.Println("Trying to Restart:", p.id)
		p.stop <- 2
		return nil
	}

	return errors.New("Not Found")
}

// Stop stops a push subscription
func (mgr *Manager) Stop(project string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(project + "/" + sub); err == nil {
		if p.running == false {
			log.Println("Already stopped", p.id, "state:", p.running)
			return errors.New("Already Stoped")
		}
		log.Println("Trying to stop:", p.id)
		p.stop <- 1
		return nil
	}

	return errors.New("Not Found")
}

// Refresh updates the subscription information from the database
func (mgr *Manager) Refresh(project string, sub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(project + "/" + sub); err == nil {
		subs := subscriptions.Subscriptions{}
		err := subs.LoadOne(project, sub, mgr.store)

		if err != nil {
			return errors.New("No sub found")
		}

		p.endpoint = subs.List[0].PushCfg.Pend
		p.retryPolicy = subs.List[0].PushCfg.RetPol.PolicyType
		p.retryPeriod = subs.List[0].PushCfg.RetPol.Period
		p.rate = time.Duration(p.retryPeriod) * time.Millisecond
	}

	return errors.New("Not Found")
}

// Add a new push subscription
func (mgr *Manager) Add(project string, subname string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}
	// Check if subscription exists
	subs := subscriptions.Subscriptions{}
	err := subs.LoadOne(project, subname, mgr.store)

	if err != nil {
		return errors.New("No sub found")
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
	mgr.list[project+"/"+subname] = &pushr
	log.Println("Push Subscription Added")

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

	return errors.New("Not Found")
}

// Launch the pusher activity
func (p *Pusher) launch(brk brokers.Broker, store stores.Store) {
	log.Println("pusher:", p.id, "launching...")
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

				log.Println("pusher:", p.id, "Stoping...")
				p.running = false
				if halt == 2 {
					p.mgr.Launch(p.sub.Project, p.sub.Name)
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
