package push

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
)

// Pusher holds information for the pusher routine and subscription
type Pusher struct {
	id       int
	sub      subscriptions.Subscription
	endpoint string
	stop     chan bool
	rate     time.Duration // in milliseconds
	running  bool
	sndr     Sender
}

// Manager manages all pusher routines
type Manager struct {
	list   map[string]*Pusher // map using as key the string = "{project}/{sub}"
	broker brokers.Broker     // Reference to backend broker
	store  stores.Store       // Reference to backend store
}

// Push method of pusher object to consume and push messages
func (p *Pusher) push(brk brokers.Broker, store stores.Store) {
	log.Println("pid", p.id, "pushing")
	// update sub details
	subs := subscriptions.Subscriptions{}
	subs.LoadOne(p.sub.Project, p.sub.Name, store.Clone())
	p.sub = subs.List[0]
	// Init Received Message List

	fullTopic := p.sub.Project + "." + p.sub.Topic
	msgs := brk.Consume(fullTopic, p.sub.Offset)
	if len(msgs) > 0 {
		err := p.sndr.Send(p.endpoint, msgs[0])
		if err != nil {
			// Advance the offset
			store.UpdateSubOffset(p.sub.Name, 1+p.sub.Offset)

		}
	} else {
		log.Println("pid:", p.id, "empty")
	}
}

// Shoutout prints manager stats
func (mgr *Manager) Shoutout() {
	for k := range mgr.list {
		item := mgr.list[k]
		log.Println("--- pid:", item.id, "running:", item.running)
	}
}

// NewManager creates a new manager object for managing push routines
func NewManager(brk brokers.Broker, str stores.Store) *Manager {
	mgr := Manager{}
	mgr.broker = brk
	mgr.store = str
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

// Stop stops a push subscription
func (mgr *Manager) Stop(psub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(psub); err == nil {
		if p.running == false {
			log.Println("Already stopped", p.id, "state:", p.running)
			return errors.New("Already Stoped")
		}
		log.Println("Trying to stop:", p.id)
		p.stop <- true
		return nil
	}

	return errors.New("Not Found")
}

// Add a new push subscription
func (mgr *Manager) Add(psub string, sndr Sender) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}
	// Check if subscription exists

	project, subname, err := splitPSub(psub)

	if err != nil {
		return err
	}

	subs := subscriptions.Subscriptions{}

	err = subs.LoadOne(project, subname, mgr.store.Clone())

	if err != nil {
		return errors.New("No sub found")
	}

	// Create new pusher

	pushr := Pusher{}
	pushr.id = len(mgr.list)
	pushr.sub = subs.List[0]
	pushr.running = false
	pushr.stop = make(chan bool, 2)
	pushr.rate = 300 * time.Millisecond
	pushr.sndr = sndr

	mgr.list[psub] = &pushr
	log.Println("Push Subscription Added")

	return nil

}

// Launch Launches a new puhser
func (mgr *Manager) Launch(psub string) error {
	// Check if mgr is set
	if !mgr.isSet() {
		return errors.New("Push Manager not set")
	}

	if p, err := mgr.Get(psub); err == nil {
		if p.running == true {
			return errors.New("Already Running")
		}
		p.launch(mgr.broker, mgr.store)
		return nil
	}

	return errors.New("Not Found")
}

// Launch the pusher activity
func (p *Pusher) launch(brk brokers.Broker, store stores.Store) {
	log.Println("pusher:", p.id, "launching...")
	p.running = true
	go Activity(p, brk, store)
}

//Activity the push subscription
func Activity(p *Pusher, brk brokers.Broker, store stores.Store) error {

	for {
		rate := time.After(p.rate)
		select {
		case <-p.stop:
			{
				log.Println("pusher:", p.id, "Stoping...")
				p.running = false
				return nil
			}
		case <-rate:
			{
				p.push(brk, store)
			}
		}
	}

}
