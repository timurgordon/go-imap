module imap

import io
import crypto.tls
import net.http.mime
import regex
import strconv
import strings
import sync
import time
import encoding
// import github.com.StirlingMarketingGroup.go-retry as retry
// import github.com.davecgh.go-spew.spew
// import github.com.dustin.go-humanize as humanize
import github.com.jhillyerd.enmime
import github.com.logrusorgru.aurora
import github.com.rs.xid
import net.html.charset

const (
	nl   = '\r\n'
	atom = regexp.must_compile('{\\d+}$')
)

// Reconnect closes the current connection (if any) and establishes a new
pub fn (mut d Dialer) reconnect() ! {
	d.close()
	if verbose {
		log(d.conn_num, d.folder, aurora.yellow(aurora.bold('reopening connection')))
	}
	mut d2, err_1 := d.clone()
	if err_1 != unsafe { nil } {
		return error(strconv.v_sprintf('imap reconnect: %s', err_1))
	}
	d = *d2
	return
}

fn drop_nl(b []u8) []u8 {
	if b.len >= 1 && b[b.len - 1] == `\n` {
		if b.len >= 2 && b[b.len - 2] == `\r` {
			return b[..b.len - 2]
		} else {
			return b[..b.len - 1]
		}
	}
	return b
}

// Exec executes the command on the imap connec
pub fn (mut d Dialer) exec(command string, buildResponse bool, retryCount int, processLine fn ([]u8) error) (string, error) {
	mut response := ''
	mut err_1 := error{}
	mut resp := strings.Builder{}
	err_1 = retry.retry(fn () error {
		mut err_2 := error{}
		mut tag := strconv.v_sprintf('%X', xid.new()).bytes()
		mut c := strconv.v_sprintf('%s %s\r\n', tag, command)
		if verbose {
			log(d.conn_num, d.folder, strings.replace(strconv.v_sprintf('%s %s', aurora.bold('->'),
				c.trim_space()), strconv.v_sprintf('"%s"', d.password), '"****"', -1))
		}
		_, err_2 = d.conn.write(c.bytes())
		if err_2 != unsafe { nil } {
			return
		}
		mut r := bufio.new_reader(d.conn)
		if buildResponse {
			resp = strings.Builder{}
		}
		mut line_1 := []u8{}
		for err_2 == unsafe { nil } {
			line_1, err_2 = r.read_bytes(`\n`)
			for {
				mut a := imap.atom.find(drop_nl(line_1))
				if a != unsafe { nil } {
					mut n := 0
					n, err_2 = strconv.atoi(a[1..a.len - 1].str())
					if err_2 != unsafe { nil } {
						return
					}
					mut buf := []u8{len: n}
					_, err_2 = io.read_full(r, buf)
					if err_2 != unsafe { nil } {
						return
					}
					line_1 << buf
					buf, err_2 = r.read_bytes(`\n`)
					if err_2 != unsafe { nil } {
						return
					}
					line_1 << buf
					continue
				}
				break
			}
			if verbose && !skip_responses {
				log(d.conn_num, d.folder, strconv.v_sprintf('<- %s', drop_nl(line_1)))
			}
			if line_1.len >= 19 && bytes.equal(line_1[..16], tag) {
				if !bytes.equal(line_1[17..19], 'OK'.bytes()) {
					err_2 = error(strconv.v_sprintf('imap command failed: %s', line_1[20..]))
					return
				}
				break
			}
			if processLine != unsafe { nil } {
				err_2 = processLine(line_1)
				if err_2 != unsafe { nil } {
					return
				}
			}
			if buildResponse {
				resp.write(line_1)
			}
		}
		return
	}, retryCount, fn (err_3 error) error {
		if verbose {
			log(d.conn_num, d.folder, aurora.red(err_3))
		}
		d.close()
		return unsafe { nil }
	}, fn () error {
		return d.reconnect()
	})
	if err_3 != unsafe { nil } {
		if verbose {
			log(d.conn_num, d.folder, aurora.red(aurora.bold('All retries failed')))
		}
		return '', err_3
	}
	if buildResponse {
		if resp.len() != 0 {
			last_resp = resp.string()
			return last_resp, unsafe { nil }
		}
		return '', unsafe { nil }
	}
	return
}

// Login attempts to l
pub fn (mut d Dialer) login(username string, password string) error {
	mut err_2 := error{}
	_, err_2 = d.exec(strconv.v_sprintf('LOGIN "%s" "%s"', add_slashes.replace(username),
		add_slashes.replace(password)), false, retry_count, unsafe { nil })
	return
}