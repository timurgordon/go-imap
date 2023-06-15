package imap

// Dialer is basically an IMAP connection
type Dialer struct {
	conn      *tls.Conn
	Folder    string
	Username  string
	Password  string
	Host      string
	Port      int
	strtokI   int
	strtok    string
	Connected bool
	ConnNum   int
}

// EmailAddresses are a map of email address to names
type EmailAddresses map[string]string

// Email is an email message
type Email struct {
	Flags       []string
	Received    time.Time
	Sent        time.Time
	Size        uint64
	Subject     string
	UID         int
	MessageID   string
	From        EmailAddresses
	To          EmailAddresses
	ReplyTo     EmailAddresses
	CC          EmailAddresses
	BCC         EmailAddresses
	Text        string
	HTML        string
	Attachments []Attachment
}

// Attachment is an Email attachment
type Attachment struct {
	Name     string
	MimeType string
	Content  []byte
}
