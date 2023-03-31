package main

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
