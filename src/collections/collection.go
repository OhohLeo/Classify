package collections

import (
	"errors"
	"fmt"
	"github.com/ohohleo/classify/exports"
	"github.com/ohohleo/classify/imports"
	"github.com/ohohleo/classify/websites"
	"log"
)

type Collection struct {
	name     string
	buffer   *Buffer
	items    *Items
	config   *Config
	events   chan Event
	websites map[string]websites.Website
	exports  []exports.Export
}

type Event struct {
	Source string
	Status string
	Id     string
	Item   Data
}

func (c *Collection) Init(name string) chan Event {

	c.SetName(name)

	// Set default Buffer size at 2
	c.config = NewConfig(2)

	// Init the items buffer
	c.ResetBuffer()

	// Init the items storage
	c.items = NewItems()

	// Init event handler
	c.events = make(chan Event)

	return c.events
}

// GetType returns the type of the collection (mandatory)
func (c *Collection) GetType() string {
	panic("collection type should be specified!")
}

// SetName set the name of the collection
func (c *Collection) SetName(name string) {
	c.name = name
}

// GetName returns the name of the collection
func (c *Collection) GetName() string {
	return c.name
}

// AddWebsite add new website
func (c *Collection) AddWebsite(website websites.Website) {

	if c.websites == nil {
		c.websites = make(map[string]websites.Website)
	}

	c.websites[website.GetName()] = website
}

// DeleteWebsite delete specified website
func (c *Collection) DeleteWebsite(name string) error {

	if _, ok := c.websites[name]; ok {

		delete(c.websites, name)
		return nil
	}

	return errors.New("no website name '" + name + "' found")
}

func (c *Collection) Search(src string, item *Item) {

	// Launch research through web
	if len(c.websites) > 0 {
		c.SearchWeb(item)

		item.MatchId = item.Websites["TMDB"][0].GetId()
	}

	// // TODO Research for best matching
	// item.Type = item.Websites["IMDB"][0].GetType()
	// item.IsMatching = 10.2
	// item.BestMatch = item.Websites["IMDB"][0]

	c.SendEvent(src, "update", item)
}

// SearchWeb launch resarch through specified websites
func (c *Collection) SearchWeb(item *Item) {

	// Get name to search
	keywords := item.WebQuery

	// For all specified websites
	for _, website := range c.websites {

		// Launch the research
		channel := website.Search(keywords)

		for {
			data, ok := <-channel
			if ok {
				item.AddWebsiteData(website.GetName(), data)
				continue
			}

			break
		}
	}
}

// OnInput handle new data to classify
func (c *Collection) OnInput(input imports.Data) *Item {

	// Create a new item
	item := NewItem()

	log.Printf("OnInput %s\n", item)

	// Add the import to the item
	item.AddImportData(input)

	// Get Cleaned name
	item.SetCleanedName(c.config.Banned, c.config.Separators)

	// Store the import to the buffer collection
	c.buffer.Add(item.GetId(), item)

	return item
}

func (c *Collection) SendEvent(src string, status string, item Data) {

	if c.events == nil {
		return
	}

	c.events <- Event{
		Source: src,
		Status: status,
		Id:     item.GetId(),
		Item:   item,
	}
}

func (c *Collection) GetBuffer() []*Item {

	if c.buffer == nil {
		return []*Item{}
	}

	return c.buffer.GetCurrentList()
}

func (c *Collection) ResetBuffer() {
	c.buffer = NewBuffer(c, c.config.BufferSize)
}

func (c *Collection) GetItems() []Data {

	if c.items == nil {
		return []Data{}
	}

	return c.items.GetCurrentList()
}

func (c *Collection) ResetItems() {
	c.items = NewItems()
}

func (c *Collection) Validate(id string, data Data) (item *Item, err error) {

	if c.buffer == nil {
		err = fmt.Errorf("buffer not initialized")
		return
	}

	item = c.buffer.Validate(id)
	if item == nil {
		err = fmt.Errorf("name '%s' not referenced in buffer", id)
		return
	}

	// Store in definitive items
	err = c.items.Add(id, data)
	if err != nil {
		return
	}

	c.SendEvent("items", "add", data)

	// TODO: export list

	return
}

func (c *Collection) SetExports(exports []exports.Export) {
	c.exports = exports
}

// OnOutput export data from classify
func (c *Collection) onOutput(item *Item) error {

	return nil
}
