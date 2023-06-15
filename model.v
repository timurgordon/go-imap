module imap

type EmailAddresses = map[string]string

struct Dialer {
mut:
	conn      &tls.Conn
	folder    string
	username  string
	password  string
	host      string
	port      int
	strtok_i  int
	strtok    string
	connected bool
	conn_num  int
}

struct Attachment {
mut:
	name      string
	mime_type string
	content   []u8
}

struct Email {
mut:
	flags       []string
	received    time.Time
	sent        time.Time
	size        u64
	subject     string
	uid         int
	message_id  string
	from        EmailAddresses
	to          EmailAddresses
	reply_to    EmailAddresses
	cc          EmailAddresses
	bcc         EmailAddresses
	text        string
	html        string
	attachments []Attachment
}
