package main

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"regexp"
	"strings"
	"time"
)

func stripPort(host string) string {
	if colon := strings.IndexByte(host, ':'); colon != -1 {
		return host[:colon]
	}
	return host
}

func dateString() string {
	return time.Now().UTC().Format(time.RFC850)
}

var re = regexp.MustCompile(`^([a-z]+).iambroken.com$`)

var mainPage = struct {
	body    string
	modtime time.Time
}{`iambroken.com offers simple tools to help you test your network-using stuff.
We return plain text with a trailing newline without ads or extra content.

If you have any suggestions or complaints, email sidney+iambroken@s4y.us.

(Thanks to Russ for ua. and localtest.me.)

-------------------------------------------------------------------------------

http://ip.iambroken.com   - Your IP address
                            (also see http://istheinternetdown.com/)

http://ua.iambroken.com   - Your user agent

http://time.iambroken.com - GMT time

http://echo.iambroken.com - Right back at you, HTTP requests
                            (must be well-formed)

loop.iambroken.com        - 127.0.0.1. Smells like home.
                            (also see http://readme.localtest.me/)
`, time.Unix(1457713257, 0)}

func basic(f func(*http.Request) string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.ServeContent(w, r, "", time.Time{}, strings.NewReader(f(r)+"\n"))
	}
}

var tools = func() map[string]func(http.ResponseWriter, *http.Request) {
	tools := make(map[string]func(http.ResponseWriter, *http.Request))
	tools["ip"] = basic(func(r *http.Request) string {
		return stripPort(r.RemoteAddr)
	})
	tools["ua"] = basic(func(r *http.Request) string {
		return r.Header.Get("User-Agent")
	})
	tools["time"] = basic(func(r *http.Request) string {
		return dateString()
	})
	tools["echo"] = func(w http.ResponseWriter, r *http.Request) {
		// Hack around Go automatically inserting a UA
		if _, ok := r.Header["User-Agent"]; !ok {
			r.Header.Set("User-Agent", "")
		}

		// Hack away NFSN headers
		r.Header.Del("Surrogate-Capability")
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("X-Forwarded-Host")
		r.Header.Del("X-Forwarded-Proto")
		r.Header.Del("X-Forwarded-Server")
		r.Header.Del("X-Nfsn-Https")
		r.Header.Del("Via")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		r.Write(w)
	}
	return tools
}()

var istheinternetdown_basic = basic(func(r *http.Request) string {
	return strings.Join([]string{
		"Everything's shiny.\n",
		dateString(),
		"Your IP address: " + stripPort(r.RemoteAddr),
	}, "\n")
})

var istheinternetdown_tmpl = template.Must(template.New("istheinternetdown").Parse(`<!DOCTYPE html>
<title>Is the internet down?</title>
<style>
	body{
		background-color: #{{.Color}};
		text-align: center;
		font: 14px "Helvetica Neue", "Helvetica", "Arial", sans-serif;
	}
	h1{
		font-weight: normal;
		font-size: 4.2em;
		margin-top: 15%;
		margin-bottom: 0.8em;
	}
</style>
<h1>Doesn&rsquo;t look like it.</h1>
<p>(You called from <code>{{.IP}}</code> on {{.Date}}.)</p>
`))

type handler func(http.ResponseWriter, *http.Request)

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(fcgi.Serve(ln, handler(func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)
		switch host {
		case "www.iambroken.com":
			http.Redirect(w, r, "//iambroken.com/", http.StatusFound)
			return
		case "www.istheinternetdown.com":
			http.Redirect(w, r, "//istheinternetdown.com/", http.StatusFound)
			return
		case "istheinternetdown.com":
			if r.URL.Path != "/" {
				break
			}
			if strings.Index(r.Header.Get("Accept"), "html") == -1 {
				istheinternetdown_basic(w, r)
			} else {
				color := make([]byte, 3)
				_, _ = rand.Read(color)
				istheinternetdown_tmpl.Execute(w, struct {
					Color string
					IP    string
					Date  string
				}{
					hex.EncodeToString(color),
					stripPort(r.RemoteAddr),
					dateString(),
				})
			}
			return
		case "iambroken.com":
			if r.URL.Path != "/" {
				break
			}
			http.ServeContent(w, r, "", mainPage.modtime, strings.NewReader(mainPage.body))
			return
		default:
			if r.URL.Path != "/" {
				break
			}
			m := re.FindStringSubmatch(host)
			if m == nil {
				break
			}
			tool := tools[m[1]]
			if tool == nil {
				break
			}
			tool(w, r)
			return
		}
		http.NotFound(w, r)
	})))
}
