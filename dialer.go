package imap

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	retry "github.com/StirlingMarketingGroup/go-retry"
	"github.com/davecgh/go-spew/spew"
	humanize "github.com/dustin/go-humanize"
	"github.com/jhillyerd/enmime"
	"github.com/logrusorgru/aurora"
	"github.com/rs/xid"
	"golang.org/x/net/html/charset"
)

// Reconnect closes the current connection (if any) and establishes a new one
func (d *Dialer) Reconnect() (err error) {
	d.Close()
	if Verbose {
		log(d.ConnNum, d.Folder, aurora.Yellow(aurora.Bold("reopening connection")))
	}
	d2, err := d.Clone()
	if err != nil {
		return fmt.Errorf("imap reconnect: %s", err)
	}
	*d = *d2
	return
}

const nl = "\r\n"

func dropNl(b []byte) []byte {
	if len(b) >= 1 && b[len(b)-1] == '\n' {
		if len(b) >= 2 && b[len(b)-2] == '\r' {
			return b[:len(b)-2]
		} else {
			return b[:len(b)-1]
		}
	}
	return b
}

var atom = regexp.MustCompile(`{\d+}$`)

// Exec executes the command on the imap connection
func (d *Dialer) Exec(command string, buildResponse bool, retryCount int, processLine func(line []byte) error) (response string, err error) {
	var resp strings.Builder
	err = retry.Retry(func() (err error) {
		tag := []byte(fmt.Sprintf("%X", xid.New()))

		c := fmt.Sprintf("%s %s\r\n", tag, command)

		if Verbose {
			log(d.ConnNum, d.Folder, strings.Replace(fmt.Sprintf("%s %s", aurora.Bold("->"), strings.TrimSpace(c)), fmt.Sprintf(`"%s"`, d.Password), `"****"`, -1))
		}

		_, err = d.conn.Write([]byte(c))
		if err != nil {
			return
		}

		r := bufio.NewReader(d.conn)

		if buildResponse {
			resp = strings.Builder{}
		}
		var line []byte
		for err == nil {
			line, err = r.ReadBytes('\n')
			for {
				if a := atom.Find(dropNl(line)); a != nil {
					// fmt.Printf("%s\n", a)
					var n int
					n, err = strconv.Atoi(string(a[1 : len(a)-1]))
					if err != nil {
						return
					}

					buf := make([]byte, n)
					_, err = io.ReadFull(r, buf)
					if err != nil {
						return
					}
					line = append(line, buf...)

					buf, err = r.ReadBytes('\n')
					if err != nil {
						return
					}
					line = append(line, buf...)

					continue
				}
				break
			}

			if Verbose && !SkipResponses {
				log(d.ConnNum, d.Folder, fmt.Sprintf("<- %s", dropNl(line)))
			}

			// if strings.Contains(string(line), "--00000000000030095105741e7f1f") {
			// 	f, _ := ioutil.TempFile("", "")
			// 	ioutil.WriteFile(f.Name(), line, 0777)
			// 	fmt.Println(f.Name())
			// }

			if len(line) >= 19 && bytes.Equal(line[:16], tag) {
				if !bytes.Equal(line[17:19], []byte("OK")) {
					err = fmt.Errorf("imap command failed: %s", line[20:])
					return
				}
				break
			}

			if processLine != nil {
				if err = processLine(line); err != nil {
					return
				}
			}
			if buildResponse {
				resp.Write(line)
			}
		}
		return
	}, retryCount, func(err error) error {
		if Verbose {
			log(d.ConnNum, d.Folder, aurora.Red(err))
		}
		d.Close()
		return nil
	}, func() error {
		return d.Reconnect()
	})
	if err != nil {
		if Verbose {
			log(d.ConnNum, d.Folder, aurora.Red(aurora.Bold("All retries failed")))
		}
		return "", err
	}

	if buildResponse {
		if resp.Len() != 0 {
			lastResp = resp.String()
			return lastResp, nil
		}
		return "", nil
	}
	return
}

// Login attempts to login
func (d *Dialer) Login(username string, password string) (err error) {
	_, err = d.Exec(fmt.Sprintf(`LOGIN "%s" "%s"`, AddSlashes.Replace(username), AddSlashes.Replace(password)), false, RetryCount, nil)
	return
}


