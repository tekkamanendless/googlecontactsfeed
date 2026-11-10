package googlecontactsfeed

import (
	"encoding/xml"
	"time"
)

type ContactFeed struct {
	XMLName     xml.Name  `xml:"feed"`
	ID          string    `xml:"id"`
	Updated     time.Time `xml:"updated"`
	Category    Category  `xml:"category"`
	Title       string    `xml:"title"`
	Link        []Link    `xml:"link"`
	AuthorName  string    `xml:"author>name"`
	AuthorEmail string    `xml:"author>email"`
	Generator   struct {
		Version string `xml:"version,attr"`
		URI     string `xml:"uri,attr"`
		Value   string `xml:",chardata"`
	} `xml:"generator"`
	OpenSearchTotalResults int             `xml:"openSearch totalResults"`
	OpenSearchStartIndex   int             `xml:"openSearch startIndex"`
	OpenSearchItemsPerPage int             `xml:"openSearch itemsPerPage"`
	Entries                []GoogleContact `xml:"entry"`
}

type GoogleContact struct {
	XMLName              xml.Name           `xml:"entry"`
	ID                   string             `xml:"id,omitempty"`
	Updated              *time.Time         `xml:"updated,omitempty"`
	Category             Category           `xml:"category"`
	Title                string             `xml:"title,omitempty"` // This is read-only.
	Links                []Link             `xml:"link"`
	Deleted              bool               `xml:"http://schemas.google.com/g/2005 deleted"` // This will be set if "showdeleted" is true and the contact was deleted.
	Name                 GDName             `xml:"http://schemas.google.com/g/2005 name"`
	GDEmails             []GDEmail          `xml:"http://schemas.google.com/g/2005 email"`
	GDPhoneNumbers       []GDPhoneNumber    `xml:"http://schemas.google.com/g/2005 phoneNumber"`
	GDExtendedProperties []ExtendedProperty `xml:"http://schemas.google.com/g/2005 extendedProperty"`
}

// EditURL returns the edit URL for the contact.
//
// This URL can be used to update or delete the contact.
func (c *GoogleContact) EditURL() string {
	for _, link := range c.Links {
		if link.Rel == "edit" {
			return link.Href
		}
	}
	return ""
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

type Category struct {
	Scheme string `xml:"id,attr"`
	Term   string `xml:"term,attr"`
}

type GDName struct {
	GivenName  string `xml:"http://schemas.google.com/g/2005 givenName"`
	FamilyName string `xml:"http://schemas.google.com/g/2005 familyName"`
	FullName   string `xml:"http://schemas.google.com/g/2005 fullName"`
}

type GDEmail struct {
	Rel         string `xml:"rel,attr"`
	Primary     bool   `xml:"primary,attr,omitempty"`
	Address     string `xml:"address,attr"`
	DisplayName string `xml:"displayName,attr,omitempty"`
}

type GDPhoneNumber struct {
	Rel     string `xml:"rel,attr"`
	Primary bool   `xml:"primary,attr,omitempty"`
	Value   string `xml:",chardata"`
}

type ExtendedProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}
