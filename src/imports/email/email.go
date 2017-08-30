package email

import (
	"encoding/json"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/ohohleo/classify/data"
	"github.com/ohohleo/classify/imports"
	"log"
	"time"
)

const (
	SEARCH Request = iota
	ALL
)

type Request int

type Email struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Login    string `json:"login"`
	Password string `json:"password"`

	Request      Request `json:"request"`
	MailBox      string  `json:"mailbox"`
	OnlyAttached bool    `json:"onlyAttached"`
	Search       Search  `json:"search"`

	dataChannel chan data.Data
	cnx         *client.Client
}

type Search struct {
	Since   time.Time `json:"since"`
	Before  time.Time `json:"before"`
	Larger  uint32    `json:"larger"`
	Smaller uint32    `json:"smaller"`
	Text    []string  `json:"text"`
}

func (s *Search) IsValid() bool {
	return true
}

type EmailOutputParams struct {
	MailBoxes []string `json:"mailboxes"`
}

func ToBuild() imports.BuildImport {
	return imports.BuildImport{
		Create: Create,
	}
}

func Create(input json.RawMessage, config map[string][]string, collections []string) (i imports.Import, params interface{}, err error) {

	var email Email
	err = json.Unmarshal(input, &email)

	fmt.Printf("%+v\n", email)

	if email.MailBox == "" {

		// Check connection
		err = email.Connect()
		if err != nil {
			return
		}

		// Returns mailbox
		var mailboxes []string
		mailboxes, err = email.GetMailBoxes()
		if err != nil {
			return
		}

		params = &EmailOutputParams{
			MailBoxes: mailboxes,
		}

		email.Stop()
		err = fmt.Errorf("import 'email' needs more params")
		return
	}

	switch email.Request {
	case SEARCH:
		if email.Search.IsValid() == false {
			err = fmt.Errorf("import 'email' invalid search params")
			return
		}
	case ALL:
	}

	i = &email

	return
}

func (e *Email) GetRef() imports.Ref {
	return imports.EMAIL
}

func (e *Email) Check(config map[string][]string, collections []string) error {
	return nil
}

func (e *Email) Start() (c chan data.Data, err error) {

	// Check if the analysis is not already going on
	if e.cnx != nil {
		err = fmt.Errorf("import 'email' already started")
		return
	}

	// Establish connection
	if err = e.Connect(); err != nil {
		return
	}

	switch e.Request {
	case SEARCH:
		go e.GetSearch()
	case ALL:
		go e.GetAllMessages()
	}

	c = make(chan data.Data)

	e.dataChannel = c

	return
}

func (e *Email) Stop() error {

	// No need to close unitialised connection
	if e.cnx == nil {
		return fmt.Errorf("import 'email' already stopped")
	}

	e.cnx.Logout()
	e.cnx = nil
	return nil
}

func (e *Email) Eq(new imports.Import) bool {
	newEmail, _ := new.(*Email)
	return e.Host == newEmail.Host &&
		e.Port == newEmail.Port &&
		e.Login == newEmail.Login
}

func (e *Email) Connect() error {

	address := fmt.Sprintf("%s:%d", e.Host, e.Port)

	log.Printf("Connecting to '%s'...\n", address)

	// Connect to IMAP server
	cnx, err := client.DialTLS(address, nil)
	if err != nil {
		return fmt.Errorf("import 'email' connection: %s", err.Error())
	}

	// Login
	if err := cnx.Login(e.Login, e.Password); err != nil {
		return fmt.Errorf("import 'email' login: %s", err.Error())
	}

	log.Printf("'%s' Connected!\n", address)

	// Store connection
	e.cnx = cnx

	return nil
}

// Permet la gestion des commandes asynchrones
func (e *Email) GetMailBoxes() (mailboxes []string, err error) {

	if e.cnx == nil {
		err = fmt.Errorf("email uninitialised")
		return
	}

	// List mailboxes
	infos := make(chan *imap.MailboxInfo, 10)
	done := make(chan error)

	go func() {
		done <- e.cnx.List("", "*", infos)
	}()

	if err = <-done; err != nil {
		e.Stop()
		return
	}

	for m := range infos {
		mailboxes = append(mailboxes, m.Name)
	}

	return
}

func (e *Email) GetAllMessages() error {

	mailbox, err := e.cnx.Select(e.MailBox, false)
	if err != nil {
		return err
	}

	// Get all messages
	from := uint32(1)
	to := mailbox.Messages

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	return e.Proceed(seqset)
}

func (e *Email) GetSearch() error {

	_, err := e.cnx.Select(e.MailBox, false)
	if err != nil {
		return err
	}

	criteria := &imap.SearchCriteria{
		Since:   e.Search.Since,
		Before:  e.Search.Before,
		Larger:  e.Search.Larger,
		Smaller: e.Search.Smaller,
		Text:    e.Search.Text,
	}

	// Launch research
	seqNums, err := e.cnx.Search(criteria)
	if err != nil {
		fmt.Printf("ERROR: %+v\n", err)
		e.Stop()
		return err
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(seqNums...)

	return e.Proceed(seqset)
}

func (e *Email) Proceed(seqset *imap.SeqSet) error {

	messages := make(chan *imap.Message, 10)
	done := make(chan error)

	go func() {
		done <- e.cnx.Fetch(seqset, []string{imap.BodyMsgAttr}, messages)
	}()

	for msg := range messages {
		fmt.Printf("%+v \n", msg.SeqNum)
	}

	if err := <-done; err != nil {
		return err
	}

	close(e.dataChannel)
	e.Stop()
	return nil
}
