package main

import "time"

type Payload struct {
	UploadOnly  bool          `json:"uploadOnly"`
	IsDraft     bool          `json:"isDraft"`
	From        From          `json:"from"`
	To          []To          `json:"to"`
	Cc          []Cc          `json:"cc"`
	Bcc         []Bcc         `json:"bcc"`
	Attachments []Attachments `json:"attachments"`
	Subject     string        `json:"subject"`
	Date        string        `json:"date"`
	Text        string        `json:"text"`
	HTML        string        `json:"html"`
}
type From struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
type To struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
type Cc struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Bcc struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
type Attachments struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type GetMailBoxesResponse struct {
	Success bool `json:"success"`
	Results []struct {
		ID          string      `json:"id"`
		Name        string      `json:"name"`
		Path        string      `json:"path"`
		SpecialUse  interface{} `json:"specialUse"`
		ModifyIndex int         `json:"modifyIndex"`
		Subscribed  bool        `json:"subscribed"`
		Hidden      bool        `json:"hidden"`
		Total       int         `json:"total"`
		Unseen      int         `json:"unseen"`
		Retention   int64       `json:"retention,omitempty"`
	} `json:"results"`
}

type ReponseMakeReadAll struct {
	Success bool `json:"success"`
	Updated int  `json:"updated"`
}

type GetAllMailBoxResponse struct {
	Success bool `json:"success"`
	Total   int  `json:"total"`
	Page    int  `json:"page"`
	// PreviousCursor bool `json:"previousCursor"`
	// NextCursor     bool `json:"nextCursor"`
	Results []struct {
		ID      int    `json:"id"`
		Mailbox string `json:"mailbox"`
		Thread  string `json:"thread"`
		From    struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"from"`
		To []struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"to"`
		Cc          []interface{} `json:"cc"`
		Bcc         []interface{} `json:"bcc"`
		MessageID   string        `json:"messageId"`
		Subject     string        `json:"subject"`
		Date        time.Time     `json:"date"`
		Idate       time.Time     `json:"idate"`
		Attachments bool          `json:"attachments"`
		Size        int           `json:"size"`
		Seen        bool          `json:"seen"`
		Deleted     bool          `json:"deleted"`
		Flagged     bool          `json:"flagged"`
		Draft       bool          `json:"draft"`
		Answered    bool          `json:"answered"`
		Forwarded   bool          `json:"forwarded"`
		References  []interface{} `json:"references"`
		ContentType struct {
			Value  string `json:"value"`
			Params struct {
				Protocol string `json:"protocol"`
				Boundary string `json:"boundary"`
			} `json:"params"`
		} `json:"contentType"`
		Encrypted bool `json:"encrypted"`
	} `json:"results"`
}
