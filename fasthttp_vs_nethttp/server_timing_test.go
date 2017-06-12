package main

import (
	//"bytes"
	"fmt"
	"io"
	//"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"github.com/valyala/fasthttp"
)

var defaultClientsCount = runtime.NumCPU()

func BenchmarkRequestCtxRedirect(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var ctx fasthttp.RequestCtx
		for pb.Next() {
			ctx.Request.SetRequestURI("http://aaa.com/fff/ss.html?sdf")
			ctx.Redirect("/foo/bar?baz=111", fasthttp.StatusFound)
		}
	})
}

func BenchmarkServerGet1ReqPerConn(b *testing.B) {
	benchmarkServerGet(b, defaultClientsCount, 1)
}

func BenchmarkServerGet2ReqPerConn(b *testing.B) {
	benchmarkServerGet(b, defaultClientsCount, 2)
}

func BenchmarkServerGet10ReqPerConn(b *testing.B) {
	benchmarkServerGet(b, defaultClientsCount, 10)
}

func BenchmarkServerGet10KReqPerConn(b *testing.B) {
	benchmarkServerGet(b, defaultClientsCount, 10000)
}

func BenchmarkNetHTTPServerGet1ReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerGet(b, defaultClientsCount, 1)
}

func BenchmarkNetHTTPServerGet2ReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerGet(b, defaultClientsCount, 2)
}

func BenchmarkNetHTTPServerGet10ReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerGet(b, defaultClientsCount, 10)
}

func BenchmarkNetHTTPServerGet10KReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerGet(b, defaultClientsCount, 10000)
}

func BenchmarkServerPost1ReqPerConn(b *testing.B) {
	benchmarkServerPost(b, defaultClientsCount, 1)
}

func BenchmarkServerPost2ReqPerConn(b *testing.B) {
	benchmarkServerPost(b, defaultClientsCount, 2)
}

func BenchmarkServerPost10ReqPerConn(b *testing.B) {
	benchmarkServerPost(b, defaultClientsCount, 10)
}

func BenchmarkServerPost10KReqPerConn(b *testing.B) {
	benchmarkServerPost(b, defaultClientsCount, 10000)
}

func BenchmarkNetHTTPServerPost1ReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerPost(b, defaultClientsCount, 1)
}

func BenchmarkNetHTTPServerPost2ReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerPost(b, defaultClientsCount, 2)
}

func BenchmarkNetHTTPServerPost10ReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerPost(b, defaultClientsCount, 10)
}

func BenchmarkNetHTTPServerPost10KReqPerConn(b *testing.B) {
	benchmarkNetHTTPServerPost(b, defaultClientsCount, 10000)
}

func BenchmarkServerGet1ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 1)
}

func BenchmarkServerGet2ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 2)
}

func BenchmarkServerGet10ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 10)
}

func BenchmarkServerGet100ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 100)
}

func BenchmarkNetHTTPServerGet1ReqPerConn10KClients(b *testing.B) {
	benchmarkNetHTTPServerGet(b, 10000, 1)
}

func BenchmarkNetHTTPServerGet2ReqPerConn10KClients(b *testing.B) {
	benchmarkNetHTTPServerGet(b, 10000, 2)
}

func BenchmarkNetHTTPServerGet10ReqPerConn10KClients(b *testing.B) {
	benchmarkNetHTTPServerGet(b, 10000, 10)
}

func BenchmarkNetHTTPServerGet100ReqPerConn10KClients(b *testing.B) {
	benchmarkNetHTTPServerGet(b, 10000, 100)
}

func BenchmarkServerHijack(b *testing.B) {
	clientsCount := 1000
	requestsPerConn := 10000
	ch := make(chan struct{}, b.N)
	responseBody := []byte("123")
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Hijack(func(c net.Conn) {
				// emulate server loop :)
				err := fasthttp.ServeConn(c, func(ctx *fasthttp.RequestCtx) {
					ctx.Success("foobar", responseBody)
					registerServedRequest(b, ch)
				})
				if err != nil {
					b.Fatalf("error when serving connection")
				}
			})
			ctx.Success("foobar", responseBody)
			registerServedRequest(b, ch)
		},
		Concurrency: 16 * clientsCount,
	}
	req := "GET /foo HTTP/1.1\r\nHost: google.com\r\n\r\n"
	benchmarkServer(b, s, clientsCount, requestsPerConn, req)
	verifyRequestsServed(b, ch)
}

func BenchmarkServerMaxConnsPerIP(b *testing.B) {
	clientsCount := 1000
	requestsPerConn := 10
	ch := make(chan struct{}, b.N)
	responseBody := []byte("123")
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Success("foobar", responseBody)
			registerServedRequest(b, ch)
		},
		MaxConnsPerIP: clientsCount * 2,
		Concurrency:   16 * clientsCount,
	}
	req := "GET /foo HTTP/1.1\r\nHost: google.com\r\n\r\n"
	benchmarkServer(b, s, clientsCount, requestsPerConn, req)
	verifyRequestsServed(b, ch)
}

func BenchmarkServerTimeoutError(b *testing.B) {
	clientsCount := 10
	requestsPerConn := 1
	ch := make(chan struct{}, b.N)
	n := uint32(0)
	responseBody := []byte("123")
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if atomic.AddUint32(&n, 1)&7 == 0 {
				ctx.TimeoutError("xxx")
				go func() {
					ctx.Success("foobar", responseBody)
				}()
			} else {
				ctx.Success("foobar", responseBody)
			}
			registerServedRequest(b, ch)
		},
		Concurrency: 16 * clientsCount,
	}
	req := "GET /foo HTTP/1.1\r\nHost: google.com\r\n\r\n"
	benchmarkServer(b, s, clientsCount, requestsPerConn, req)
	verifyRequestsServed(b, ch)
}

type fakeServerConn struct {
	net.TCPConn
	ln            *fakeListener
	requestsCount int
	pos           int
	closed        uint32
}

func (c *fakeServerConn) Read(b []byte) (int, error) {
	nn := 0
	reqLen := len(c.ln.request)
	for len(b) > 0 {
		if c.requestsCount == 0 {
			if nn == 0 {
				return 0, io.EOF
			}
			return nn, nil
		}
		pos := c.pos % reqLen
		n := copy(b, c.ln.request[pos:])
		b = b[n:]
		nn += n
		c.pos += n
		if n+pos == reqLen {
			c.requestsCount--
		}
	}
	return nn, nil
}

func (c *fakeServerConn) Write(b []byte) (int, error) {
	return len(b), nil
}

var fakeAddr = net.TCPAddr{
	IP:   []byte{1, 2, 3, 4},
	Port: 12345,
}

func (c *fakeServerConn) RemoteAddr() net.Addr {
	return &fakeAddr
}

func (c *fakeServerConn) Close() error {
	if atomic.AddUint32(&c.closed, 1) == 1 {
		c.ln.ch <- c
	}
	return nil
}

func (c *fakeServerConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *fakeServerConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type fakeListener struct {
	lock            sync.Mutex
	requestsCount   int
	requestsPerConn int
	request         []byte
	ch              chan *fakeServerConn
	done            chan struct{}
	closed          bool
}

func (ln *fakeListener) Accept() (net.Conn, error) {
	ln.lock.Lock()
	if ln.requestsCount == 0 {
		ln.lock.Unlock()
		for len(ln.ch) < cap(ln.ch) {
			time.Sleep(10 * time.Millisecond)
		}
		ln.lock.Lock()
		if !ln.closed {
			close(ln.done)
			ln.closed = true
		}
		ln.lock.Unlock()
		return nil, io.EOF
	}
	requestsCount := ln.requestsPerConn
	if requestsCount > ln.requestsCount {
		requestsCount = ln.requestsCount
	}
	ln.requestsCount -= requestsCount
	ln.lock.Unlock()
	
	c := <-ln.ch
	c.requestsCount = requestsCount
	c.closed = 0
	c.pos = 0
	
	return c, nil
}

func (ln *fakeListener) Close() error {
	return nil
}

func (ln *fakeListener) Addr() net.Addr {
	return &fakeAddr
}

func newFakeListener(requestsCount, clientsCount, requestsPerConn int, request string) *fakeListener {
	ln := &fakeListener{
		requestsCount:   requestsCount,
		requestsPerConn: requestsPerConn,
		request:         []byte(request),
		ch:              make(chan *fakeServerConn, clientsCount),
		done:            make(chan struct{}),
	}
	for i := 0; i < clientsCount; i++ {
		ln.ch <- &fakeServerConn{
			ln: ln,
		}
	}
	return ln
}

var (
	//fakeResponse = []byte("Hello, world!")
	fakeResponse = []byte(`[{	"bidder_id": 43,	"creative_id": "183840587",	"adtype": "banner",	"adcode": "<SCRIPT language='JavaScript1.1' SRC=\"https://ad.doubleclick.net/ddm/adj/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145951466;click=http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3D;?\">\n</SCRIPT>\n<NOSCRIPT>\n<A HREF=\"http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3Dhttps://ad.doubleclick.net/ddm/jump/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145961517?\">\n<IMG SRC=\"https://ad.doubleclick.net/ddm/ad/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145966533?\" BORDER=0 WIDTH=300 HEIGHT=250 ALT=\"Advertisement\"></A>\n</NOSCRIPT>\n\n<script src=\"https://c.betrad.com/surly.js#;ad_w=300;ad_h=250;coid=650;nid=17649;\" type=\"text/javascript\"></script>\n<script type=\"text/javascript\">var adloox_pc_viewed=0.75,adloox_time_viewed=3,idb=\"https%3A%2F%2Fpix.impdesk.com%2Finc%3Fm%3D--M%26a%3D-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmEX06OF0Q4y8uQg\",adloox_tracked_pixel=idb.replace('--M','ci_view'),adloox_iab_pixel=idb.replace('--M','view'),adloox_custom_pixel=idb.replace('--M','e_view'),adloox_nv_pixel=idb.replace('--M','n_view'),tab_adloox_alerte_id_infectious=['WGo48eKZAw_IbxHl','61704','300x250','302','53732','360728','openx_537145107','openx_537299375/openx_538659077','680995','usa'];</script><script type=\"text/javascript\" src=\"//j.adlooxtracking.com/ads/js/tfav_infectiousg_banoneinf.js\"></script><img src=\"https://us-east.bidswitch.net/imp/AAABWV7ubmZxVwfRs9EGz4lDcxH2LwcZceeP3A/BSWhttps_A_B_Bpix.impdesk.com_Bimp_Cp_R_I_WAUCTION__PRICE_X_Ja_R-jEzoOcmgbtCNAgEM1JjhyUBmxnIfQHPQoInscb6U2O8Y00KZG7u6-rKczyDfRk__NDgL7BWJS3ggUKIJOxG1xIAZeWGdmdonI1PFRIuvkQO2dR01AzMJtjzNYvqVgpVc__g6AANU2BgFOsQZssS58gYfsLH6q1kC1vDOMusQD9LJ7fUHtNDi9Ubq__mc1Q2v7jkSAk__TeLPpHZdOAjJxQ9KjFN__CZIYjO-hsuPpayZV555v4JT6CPYJKrC20I5xqNZpeSNW3VxsDDhE-b8umn4lMudph027w8YuWMf5TypLHW0Fig8zlhqLTsOwQ4ZvTXgtFgTy1H1yBCo6N8i-lejpQ__eYBSEjfmbgEDGs__3J8zY5q2__yzxlM__ddhRODjspJAhDfSLaesTO2zmQEwdaz4sqQ6_Jx_R302/zCJNKWDwsvvPZOkGBAhgxQ_gndE2v6TjVAruns-J-GeNQN_MA6Lre7UhrpAxYSS3c4eIMN5oxGCx6Qt_gI6r3W7UQKQHoCCbjplbTclGm0FohifYBb4NY4kNQZYq1hyxoJvl6BvpZSXI2nHUlO7cqLWqMeSM3d4qsKRm_HLsyvzo1C17sHkZd5qVadQrNWJsMs34VeFf-tD8gpsWJB5mJLx3Y732OrfLlQbD4z_WEjGrleygq7Nmv30JsgM1yewA-zibLOjnHiEXtfJe_Pv8ldNIlr286LxjTYtU3afBgqSqKnI7eusKCqfI7M1fx6M3CPPOOC2nTlYbpz8nv4goDANjUW6WT1CYTn11TKWgPTVMatNmkMRGMFdL99Wrc0jJkcVFByXyZbqpFQv8n-Ny6oPdOVH8OunQK5dA3iPsZC9uvRzAnN1PNdUOPAPg55mrAozcwCAck5VwBN0XDTGe4_ZcqpNYJCXn2h693OyVYvTeo3p1B2C8pCV6KclwFK0Op4_SHLldvhRajyQ/\" alt=\" \" style=\"display:none\"/><img src=\"https://us-east-sync.bidswitch.net/sync?ssp=openx&amp;dsp_id=25&amp;imp=1\" alt=\" \" style=\"display:none\"/>\n<iframe src='https://us-u.openx.net/w/1.0/pd?plm=5&ph=3d7a7809-a65a-442d-bf4e-f9613786ff87' width='0' height='0' style='display:none;'></iframe>  <div id='beacon_59859' style='position:absolute;left:0px;top:0px;visibility:hidden;'>\n    <img src='https://rtb-xv.openx.net/win/medianet?p=0.25&t=1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ&c=USD&s=1'/>\n  </div>",	"publisher_id": "new-default",	"tp": "http://qsearch-a.akamaihd.net/log?logid=kfk&evtid=rtbstl&url=http%3A%2F%2Fhttp%253A%252F%252Fdigitach.net%252F&domain=http&size=300x250&ext_user_id=0&iid=0&auction_id=7b54575cd51c4286921bec5275301b54&ip=172.16.200.248&crid=183840587&b43b=10.0&bc=0.25&rtbW=43&app_dis=0&cc=RU&bname=&bid_type=-1&bmlevel=0<cm_win_macro>",	"size": "300x250",	"h": 250,	"w": 300,	"server_extras": {		"bid": 0.25,		"bidder_id": 43,		"og_bid": 10	},	"instl": 1,	"creative_type": "html",	"bidderInfoBean": {		"responseTime": 130	},	"view_width": 300,	"view_height": 250}, {	"bidder_id": 43,	"creative_id": "183840587",	"adtype": "banner",	"adcode": "<SCRIPT language='JavaScript1.1' SRC=\"https://ad.doubleclick.net/ddm/adj/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145951466;click=http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3D;?\">\n</SCRIPT>\n<NOSCRIPT>\n<A HREF=\"http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3Dhttps://ad.doubleclick.net/ddm/jump/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145961517?\">\n<IMG SRC=\"https://ad.doubleclick.net/ddm/ad/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145966533?\" BORDER=0 WIDTH=300 HEIGHT=250 ALT=\"Advertisement\"></A>\n</NOSCRIPT>\n\n<script src=\"https://c.betrad.com/surly.js#;ad_w=300;ad_h=250;coid=650;nid=17649;\" type=\"text/javascript\"></script>\n<script type=\"text/javascript\">var adloox_pc_viewed=0.75,adloox_time_viewed=3,idb=\"https%3A%2F%2Fpix.impdesk.com%2Finc%3Fm%3D--M%26a%3D-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmEX06OF0Q4y8uQg\",adloox_tracked_pixel=idb.replace('--M','ci_view'),adloox_iab_pixel=idb.replace('--M','view'),adloox_custom_pixel=idb.replace('--M','e_view'),adloox_nv_pixel=idb.replace('--M','n_view'),tab_adloox_alerte_id_infectious=['WGo48eKZAw_IbxHl','61704','300x250','302','53732','360728','openx_537145107','openx_537299375/openx_538659077','680995','usa'];</script><script type=\"text/javascript\" src=\"//j.adlooxtracking.com/ads/js/tfav_infectiousg_banoneinf.js\"></script><img src=\"https://us-east.bidswitch.net/imp/AAABWV7ubmZxVwfRs9EGz4lDcxH2LwcZceeP3A/BSWhttps_A_B_Bpix.impdesk.com_Bimp_Cp_R_I_WAUCTION__PRICE_X_Ja_R-jEzoOcmgbtCNAgEM1JjhyUBmxnIfQHPQoInscb6U2O8Y00KZG7u6-rKczyDfRk__NDgL7BWJS3ggUKIJOxG1xIAZeWGdmdonI1PFRIuvkQO2dR01AzMJtjzNYvqVgpVc__g6AANU2BgFOsQZssS58gYfsLH6q1kC1vDOMusQD9LJ7fUHtNDi9Ubq__mc1Q2v7jkSAk__TeLPpHZdOAjJxQ9KjFN__CZIYjO-hsuPpayZV555v4JT6CPYJKrC20I5xqNZpeSNW3VxsDDhE-b8umn4lMudph027w8YuWMf5TypLHW0Fig8zlhqLTsOwQ4ZvTXgtFgTy1H1yBCo6N8i-lejpQ__eYBSEjfmbgEDGs__3J8zY5q2__yzxlM__ddhRODjspJAhDfSLaesTO2zmQEwdaz4sqQ6_Jx_R302/zCJNKWDwsvvPZOkGBAhgxQ_gndE2v6TjVAruns-J-GeNQN_MA6Lre7UhrpAxYSS3c4eIMN5oxGCx6Qt_gI6r3W7UQKQHoCCbjplbTclGm0FohifYBb4NY4kNQZYq1hyxoJvl6BvpZSXI2nHUlO7cqLWqMeSM3d4qsKRm_HLsyvzo1C17sHkZd5qVadQrNWJsMs34VeFf-tD8gpsWJB5mJLx3Y732OrfLlQbD4z_WEjGrleygq7Nmv30JsgM1yewA-zibLOjnHiEXtfJe_Pv8ldNIlr286LxjTYtU3afBgqSqKnI7eusKCqfI7M1fx6M3CPPOOC2nTlYbpz8nv4goDANjUW6WT1CYTn11TKWgPTVMatNmkMRGMFdL99Wrc0jJkcVFByXyZbqpFQv8n-Ny6oPdOVH8OunQK5dA3iPsZC9uvRzAnN1PNdUOPAPg55mrAozcwCAck5VwBN0XDTGe4_ZcqpNYJCXn2h693OyVYvTeo3p1B2C8pCV6KclwFK0Op4_SHLldvhRajyQ/\" alt=\" \" style=\"display:none\"/><img src=\"https://us-east-sync.bidswitch.net/sync?ssp=openx&amp;dsp_id=25&amp;imp=1\" alt=\" \" style=\"display:none\"/>\n<iframe src='https://us-u.openx.net/w/1.0/pd?plm=5&ph=3d7a7809-a65a-442d-bf4e-f9613786ff87' width='0' height='0' style='display:none;'></iframe>  <div id='beacon_59859' style='position:absolute;left:0px;top:0px;visibility:hidden;'>\n    <img src='https://rtb-xv.openx.net/win/medianet?p=0.25&t=1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ&c=USD&s=1'/>\n  </div>",	"publisher_id": "new-default",	"tp": "http://qsearch-a.akamaihd.net/log?logid=kfk&evtid=rtbstl&url=http%3A%2F%2Fhttp%253A%252F%252Fdigitach.net%252F&domain=http&size=300x250&ext_user_id=0&iid=0&auction_id=7b54575cd51c4286921bec5275301b54&ip=172.16.200.248&crid=183840587&b43b=10.0&bc=0.25&rtbW=43&app_dis=0&cc=RU&bname=&bid_type=-1&bmlevel=0<cm_win_macro>",	"size": "300x250",	"h": 250,	"w": 300,	"server_extras": {		"bid": 0.25,		"bidder_id": 43,		"og_bid": 10	},	"instl": 1,	"creative_type": "html",	"bidderInfoBean": {		"responseTime": 130	},	"view_width": 300,	"view_height": 250}, {	"bidder_id": 43,	"creative_id": "183840587",	"adtype": "banner",	"adcode": "<SCRIPT language='JavaScript1.1' SRC=\"https://ad.doubleclick.net/ddm/adj/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145951466;click=http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3D;?\">\n</SCRIPT>\n<NOSCRIPT>\n<A HREF=\"http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3Dhttps://ad.doubleclick.net/ddm/jump/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145961517?\">\n<IMG SRC=\"https://ad.doubleclick.net/ddm/ad/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145966533?\" BORDER=0 WIDTH=300 HEIGHT=250 ALT=\"Advertisement\"></A>\n</NOSCRIPT>\n\n<script src=\"https://c.betrad.com/surly.js#;ad_w=300;ad_h=250;coid=650;nid=17649;\" type=\"text/javascript\"></script>\n<script type=\"text/javascript\">var adloox_pc_viewed=0.75,adloox_time_viewed=3,idb=\"https%3A%2F%2Fpix.impdesk.com%2Finc%3Fm%3D--M%26a%3D-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmEX06OF0Q4y8uQg\",adloox_tracked_pixel=idb.replace('--M','ci_view'),adloox_iab_pixel=idb.replace('--M','view'),adloox_custom_pixel=idb.replace('--M','e_view'),adloox_nv_pixel=idb.replace('--M','n_view'),tab_adloox_alerte_id_infectious=['WGo48eKZAw_IbxHl','61704','300x250','302','53732','360728','openx_537145107','openx_537299375/openx_538659077','680995','usa'];</script><script type=\"text/javascript\" src=\"//j.adlooxtracking.com/ads/js/tfav_infectiousg_banoneinf.js\"></script><img src=\"https://us-east.bidswitch.net/imp/AAABWV7ubmZxVwfRs9EGz4lDcxH2LwcZceeP3A/BSWhttps_A_B_Bpix.impdesk.com_Bimp_Cp_R_I_WAUCTION__PRICE_X_Ja_R-jEzoOcmgbtCNAgEM1JjhyUBmxnIfQHPQoInscb6U2O8Y00KZG7u6-rKczyDfRk__NDgL7BWJS3ggUKIJOxG1xIAZeWGdmdonI1PFRIuvkQO2dR01AzMJtjzNYvqVgpVc__g6AANU2BgFOsQZssS58gYfsLH6q1kC1vDOMusQD9LJ7fUHtNDi9Ubq__mc1Q2v7jkSAk__TeLPpHZdOAjJxQ9KjFN__CZIYjO-hsuPpayZV555v4JT6CPYJKrC20I5xqNZpeSNW3VxsDDhE-b8umn4lMudph027w8YuWMf5TypLHW0Fig8zlhqLTsOwQ4ZvTXgtFgTy1H1yBCo6N8i-lejpQ__eYBSEjfmbgEDGs__3J8zY5q2__yzxlM__ddhRODjspJAhDfSLaesTO2zmQEwdaz4sqQ6_Jx_R302/zCJNKWDwsvvPZOkGBAhgxQ_gndE2v6TjVAruns-J-GeNQN_MA6Lre7UhrpAxYSS3c4eIMN5oxGCx6Qt_gI6r3W7UQKQHoCCbjplbTclGm0FohifYBb4NY4kNQZYq1hyxoJvl6BvpZSXI2nHUlO7cqLWqMeSM3d4qsKRm_HLsyvzo1C17sHkZd5qVadQrNWJsMs34VeFf-tD8gpsWJB5mJLx3Y732OrfLlQbD4z_WEjGrleygq7Nmv30JsgM1yewA-zibLOjnHiEXtfJe_Pv8ldNIlr286LxjTYtU3afBgqSqKnI7eusKCqfI7M1fx6M3CPPOOC2nTlYbpz8nv4goDANjUW6WT1CYTn11TKWgPTVMatNmkMRGMFdL99Wrc0jJkcVFByXyZbqpFQv8n-Ny6oPdOVH8OunQK5dA3iPsZC9uvRzAnN1PNdUOPAPg55mrAozcwCAck5VwBN0XDTGe4_ZcqpNYJCXn2h693OyVYvTeo3p1B2C8pCV6KclwFK0Op4_SHLldvhRajyQ/\" alt=\" \" style=\"display:none\"/><img src=\"https://us-east-sync.bidswitch.net/sync?ssp=openx&amp;dsp_id=25&amp;imp=1\" alt=\" \" style=\"display:none\"/>\n<iframe src='https://us-u.openx.net/w/1.0/pd?plm=5&ph=3d7a7809-a65a-442d-bf4e-f9613786ff87' width='0' height='0' style='display:none;'></iframe>  <div id='beacon_59859' style='position:absolute;left:0px;top:0px;visibility:hidden;'>\n    <img src='https://rtb-xv.openx.net/win/medianet?p=0.25&t=1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ&c=USD&s=1'/>\n  </div>",	"publisher_id": "new-default",	"tp": "http://qsearch-a.akamaihd.net/log?logid=kfk&evtid=rtbstl&url=http%3A%2F%2Fhttp%253A%252F%252Fdigitach.net%252F&domain=http&size=300x250&ext_user_id=0&iid=0&auction_id=7b54575cd51c4286921bec5275301b54&ip=172.16.200.248&crid=183840587&b43b=10.0&bc=0.25&rtbW=43&app_dis=0&cc=RU&bname=&bid_type=-1&bmlevel=0<cm_win_macro>",	"size": "300x250",	"h": 250,	"w": 300,	"server_extras": {		"bid": 0.25,		"bidder_id": 43,		"og_bid": 10	},	"instl": 1,	"creative_type": "html",	"bidderInfoBean": {		"responseTime": 130	},	"view_width": 300,	"view_height": 250},{	"bidder_id": 43,	"creative_id": "183840587",	"adtype": "banner",	"adcode": "<SCRIPT language='JavaScript1.1' SRC=\"https://ad.doubleclick.net/ddm/adj/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145951466;click=http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3D;?\">\n</SCRIPT>\n<NOSCRIPT>\n<A HREF=\"http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3Dhttps://ad.doubleclick.net/ddm/jump/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145961517?\">\n<IMG SRC=\"https://ad.doubleclick.net/ddm/ad/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145966533?\" BORDER=0 WIDTH=300 HEIGHT=250 ALT=\"Advertisement\"></A>\n</NOSCRIPT>\n\n<script src=\"https://c.betrad.com/surly.js#;ad_w=300;ad_h=250;coid=650;nid=17649;\" type=\"text/javascript\"></script>\n<script type=\"text/javascript\">var adloox_pc_viewed=0.75,adloox_time_viewed=3,idb=\"https%3A%2F%2Fpix.impdesk.com%2Finc%3Fm%3D--M%26a%3D-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmEX06OF0Q4y8uQg\",adloox_tracked_pixel=idb.replace('--M','ci_view'),adloox_iab_pixel=idb.replace('--M','view'),adloox_custom_pixel=idb.replace('--M','e_view'),adloox_nv_pixel=idb.replace('--M','n_view'),tab_adloox_alerte_id_infectious=['WGo48eKZAw_IbxHl','61704','300x250','302','53732','360728','openx_537145107','openx_537299375/openx_538659077','680995','usa'];</script><script type=\"text/javascript\" src=\"//j.adlooxtracking.com/ads/js/tfav_infectiousg_banoneinf.js\"></script><img src=\"https://us-east.bidswitch.net/imp/AAABWV7ubmZxVwfRs9EGz4lDcxH2LwcZceeP3A/BSWhttps_A_B_Bpix.impdesk.com_Bimp_Cp_R_I_WAUCTION__PRICE_X_Ja_R-jEzoOcmgbtCNAgEM1JjhyUBmxnIfQHPQoInscb6U2O8Y00KZG7u6-rKczyDfRk__NDgL7BWJS3ggUKIJOxG1xIAZeWGdmdonI1PFRIuvkQO2dR01AzMJtjzNYvqVgpVc__g6AANU2BgFOsQZssS58gYfsLH6q1kC1vDOMusQD9LJ7fUHtNDi9Ubq__mc1Q2v7jkSAk__TeLPpHZdOAjJxQ9KjFN__CZIYjO-hsuPpayZV555v4JT6CPYJKrC20I5xqNZpeSNW3VxsDDhE-b8umn4lMudph027w8YuWMf5TypLHW0Fig8zlhqLTsOwQ4ZvTXgtFgTy1H1yBCo6N8i-lejpQ__eYBSEjfmbgEDGs__3J8zY5q2__yzxlM__ddhRODjspJAhDfSLaesTO2zmQEwdaz4sqQ6_Jx_R302/zCJNKWDwsvvPZOkGBAhgxQ_gndE2v6TjVAruns-J-GeNQN_MA6Lre7UhrpAxYSS3c4eIMN5oxGCx6Qt_gI6r3W7UQKQHoCCbjplbTclGm0FohifYBb4NY4kNQZYq1hyxoJvl6BvpZSXI2nHUlO7cqLWqMeSM3d4qsKRm_HLsyvzo1C17sHkZd5qVadQrNWJsMs34VeFf-tD8gpsWJB5mJLx3Y732OrfLlQbD4z_WEjGrleygq7Nmv30JsgM1yewA-zibLOjnHiEXtfJe_Pv8ldNIlr286LxjTYtU3afBgqSqKnI7eusKCqfI7M1fx6M3CPPOOC2nTlYbpz8nv4goDANjUW6WT1CYTn11TKWgPTVMatNmkMRGMFdL99Wrc0jJkcVFByXyZbqpFQv8n-Ny6oPdOVH8OunQK5dA3iPsZC9uvRzAnN1PNdUOPAPg55mrAozcwCAck5VwBN0XDTGe4_ZcqpNYJCXn2h693OyVYvTeo3p1B2C8pCV6KclwFK0Op4_SHLldvhRajyQ/\" alt=\" \" style=\"display:none\"/><img src=\"https://us-east-sync.bidswitch.net/sync?ssp=openx&amp;dsp_id=25&amp;imp=1\" alt=\" \" style=\"display:none\"/>\n<iframe src='https://us-u.openx.net/w/1.0/pd?plm=5&ph=3d7a7809-a65a-442d-bf4e-f9613786ff87' width='0' height='0' style='display:none;'></iframe>  <div id='beacon_59859' style='position:absolute;left:0px;top:0px;visibility:hidden;'>\n    <img src='https://rtb-xv.openx.net/win/medianet?p=0.25&t=1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ&c=USD&s=1'/>\n  </div>",	"publisher_id": "new-default",	"tp": "http://qsearch-a.akamaihd.net/log?logid=kfk&evtid=rtbstl&url=http%3A%2F%2Fhttp%253A%252F%252Fdigitach.net%252F&domain=http&size=300x250&ext_user_id=0&iid=0&auction_id=7b54575cd51c4286921bec5275301b54&ip=172.16.200.248&crid=183840587&b43b=10.0&bc=0.25&rtbW=43&app_dis=0&cc=RU&bname=&bid_type=-1&bmlevel=0<cm_win_macro>",	"size": "300x250",	"h": 250,	"w": 300,	"server_extras": {		"bid": 0.25,		"bidder_id": 43,		"og_bid": 10	},	"instl": 1,	"creative_type": "html",	"bidderInfoBean": {		"responseTime": 130	},	"view_width": 300,	"view_height": 250},{	"bidder_id": 43,	"creative_id": "183840587",	"adtype": "banner",	"adcode": "<SCRIPT language='JavaScript1.1' SRC=\"https://ad.doubleclick.net/ddm/adj/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145951466;click=http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3D;?\">\n</SCRIPT>\n<NOSCRIPT>\n<A HREF=\"http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3Dhttps://ad.doubleclick.net/ddm/jump/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145961517?\">\n<IMG SRC=\"https://ad.doubleclick.net/ddm/ad/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145966533?\" BORDER=0 WIDTH=300 HEIGHT=250 ALT=\"Advertisement\"></A>\n</NOSCRIPT>\n\n<script src=\"https://c.betrad.com/surly.js#;ad_w=300;ad_h=250;coid=650;nid=17649;\" type=\"text/javascript\"></script>\n<script type=\"text/javascript\">var adloox_pc_viewed=0.75,adloox_time_viewed=3,idb=\"https%3A%2F%2Fpix.impdesk.com%2Finc%3Fm%3D--M%26a%3D-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmEX06OF0Q4y8uQg\",adloox_tracked_pixel=idb.replace('--M','ci_view'),adloox_iab_pixel=idb.replace('--M','view'),adloox_custom_pixel=idb.replace('--M','e_view'),adloox_nv_pixel=idb.replace('--M','n_view'),tab_adloox_alerte_id_infectious=['WGo48eKZAw_IbxHl','61704','300x250','302','53732','360728','openx_537145107','openx_537299375/openx_538659077','680995','usa'];</script><script type=\"text/javascript\" src=\"//j.adlooxtracking.com/ads/js/tfav_infectiousg_banoneinf.js\"></script><img src=\"https://us-east.bidswitch.net/imp/AAABWV7ubmZxVwfRs9EGz4lDcxH2LwcZceeP3A/BSWhttps_A_B_Bpix.impdesk.com_Bimp_Cp_R_I_WAUCTION__PRICE_X_Ja_R-jEzoOcmgbtCNAgEM1JjhyUBmxnIfQHPQoInscb6U2O8Y00KZG7u6-rKczyDfRk__NDgL7BWJS3ggUKIJOxG1xIAZeWGdmdonI1PFRIuvkQO2dR01AzMJtjzNYvqVgpVc__g6AANU2BgFOsQZssS58gYfsLH6q1kC1vDOMusQD9LJ7fUHtNDi9Ubq__mc1Q2v7jkSAk__TeLPpHZdOAjJxQ9KjFN__CZIYjO-hsuPpayZV555v4JT6CPYJKrC20I5xqNZpeSNW3VxsDDhE-b8umn4lMudph027w8YuWMf5TypLHW0Fig8zlhqLTsOwQ4ZvTXgtFgTy1H1yBCo6N8i-lejpQ__eYBSEjfmbgEDGs__3J8zY5q2__yzxlM__ddhRODjspJAhDfSLaesTO2zmQEwdaz4sqQ6_Jx_R302/zCJNKWDwsvvPZOkGBAhgxQ_gndE2v6TjVAruns-J-GeNQN_MA6Lre7UhrpAxYSS3c4eIMN5oxGCx6Qt_gI6r3W7UQKQHoCCbjplbTclGm0FohifYBb4NY4kNQZYq1hyxoJvl6BvpZSXI2nHUlO7cqLWqMeSM3d4qsKRm_HLsyvzo1C17sHkZd5qVadQrNWJsMs34VeFf-tD8gpsWJB5mJLx3Y732OrfLlQbD4z_WEjGrleygq7Nmv30JsgM1yewA-zibLOjnHiEXtfJe_Pv8ldNIlr286LxjTYtU3afBgqSqKnI7eusKCqfI7M1fx6M3CPPOOC2nTlYbpz8nv4goDANjUW6WT1CYTn11TKWgPTVMatNmkMRGMFdL99Wrc0jJkcVFByXyZbqpFQv8n-Ny6oPdOVH8OunQK5dA3iPsZC9uvRzAnN1PNdUOPAPg55mrAozcwCAck5VwBN0XDTGe4_ZcqpNYJCXn2h693OyVYvTeo3p1B2C8pCV6KclwFK0Op4_SHLldvhRajyQ/\" alt=\" \" style=\"display:none\"/><img src=\"https://us-east-sync.bidswitch.net/sync?ssp=openx&amp;dsp_id=25&amp;imp=1\" alt=\" \" style=\"display:none\"/>\n<iframe src='https://us-u.openx.net/w/1.0/pd?plm=5&ph=3d7a7809-a65a-442d-bf4e-f9613786ff87' width='0' height='0' style='display:none;'></iframe>  <div id='beacon_59859' style='position:absolute;left:0px;top:0px;visibility:hidden;'>\n    <img src='https://rtb-xv.openx.net/win/medianet?p=0.25&t=1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ&c=USD&s=1'/>\n  </div>",	"publisher_id": "new-default",	"tp": "http://qsearch-a.akamaihd.net/log?logid=kfk&evtid=rtbstl&url=http%3A%2F%2Fhttp%253A%252F%252Fdigitach.net%252F&domain=http&size=300x250&ext_user_id=0&iid=0&auction_id=7b54575cd51c4286921bec5275301b54&ip=172.16.200.248&crid=183840587&b43b=10.0&bc=0.25&rtbW=43&app_dis=0&cc=RU&bname=&bid_type=-1&bmlevel=0<cm_win_macro>",	"size": "300x250",	"h": 250,	"w": 300,	"server_extras": {		"bid": 0.25,		"bidder_id": 43,		"og_bid": 10	},	"instl": 1,	"creative_type": "html",	"bidderInfoBean": {		"responseTime": 130	},	"view_width": 300,	"view_height": 250},{	"bidder_id": 43,	"creative_id": "183840587",	"adtype": "banner",	"adcode": "<SCRIPT language='JavaScript1.1' SRC=\"https://ad.doubleclick.net/ddm/adj/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145951466;click=http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3D;?\">\n</SCRIPT>\n<NOSCRIPT>\n<A HREF=\"http://pix.impdesk.com/click?a=-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmdjs0_kkqulo_JkkrFQdU5ELSA2Z6a6n2Vx01zPxuaOR8IFOrHA&redirect=https%3A%2F%2Fnytimes-d.openx.net%2Fw%2F1.0%2Frc%3Fts%3D1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ%26r%3Dhttps://ad.doubleclick.net/ddm/jump/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145961517?\">\n<IMG SRC=\"https://ad.doubleclick.net/ddm/ad/N342209.2180301INFECTIOUSMEDIA-I/B10727729.143967290;sz=300x250;kw=IDB53732,IDB61704;ord=1483356401145966533?\" BORDER=0 WIDTH=300 HEIGHT=250 ALT=\"Advertisement\"></A>\n</NOSCRIPT>\n\n<script src=\"https://c.betrad.com/surly.js#;ad_w=300;ad_h=250;coid=650;nid=17649;\" type=\"text/javascript\"></script>\n<script type=\"text/javascript\">var adloox_pc_viewed=0.75,adloox_time_viewed=3,idb=\"https%3A%2F%2Fpix.impdesk.com%2Finc%3Fm%3D--M%26a%3D-iYzoOcmgbtCNAgEM1JjhyUimwFmfyPDHqgNfi1O1y9ufUEmEX06OF0Q4y8uQg\",adloox_tracked_pixel=idb.replace('--M','ci_view'),adloox_iab_pixel=idb.replace('--M','view'),adloox_custom_pixel=idb.replace('--M','e_view'),adloox_nv_pixel=idb.replace('--M','n_view'),tab_adloox_alerte_id_infectious=['WGo48eKZAw_IbxHl','61704','300x250','302','53732','360728','openx_537145107','openx_537299375/openx_538659077','680995','usa'];</script><script type=\"text/javascript\" src=\"//j.adlooxtracking.com/ads/js/tfav_infectiousg_banoneinf.js\"></script><img src=\"https://us-east.bidswitch.net/imp/AAABWV7ubmZxVwfRs9EGz4lDcxH2LwcZceeP3A/BSWhttps_A_B_Bpix.impdesk.com_Bimp_Cp_R_I_WAUCTION__PRICE_X_Ja_R-jEzoOcmgbtCNAgEM1JjhyUBmxnIfQHPQoInscb6U2O8Y00KZG7u6-rKczyDfRk__NDgL7BWJS3ggUKIJOxG1xIAZeWGdmdonI1PFRIuvkQO2dR01AzMJtjzNYvqVgpVc__g6AANU2BgFOsQZssS58gYfsLH6q1kC1vDOMusQD9LJ7fUHtNDi9Ubq__mc1Q2v7jkSAk__TeLPpHZdOAjJxQ9KjFN__CZIYjO-hsuPpayZV555v4JT6CPYJKrC20I5xqNZpeSNW3VxsDDhE-b8umn4lMudph027w8YuWMf5TypLHW0Fig8zlhqLTsOwQ4ZvTXgtFgTy1H1yBCo6N8i-lejpQ__eYBSEjfmbgEDGs__3J8zY5q2__yzxlM__ddhRODjspJAhDfSLaesTO2zmQEwdaz4sqQ6_Jx_R302/zCJNKWDwsvvPZOkGBAhgxQ_gndE2v6TjVAruns-J-GeNQN_MA6Lre7UhrpAxYSS3c4eIMN5oxGCx6Qt_gI6r3W7UQKQHoCCbjplbTclGm0FohifYBb4NY4kNQZYq1hyxoJvl6BvpZSXI2nHUlO7cqLWqMeSM3d4qsKRm_HLsyvzo1C17sHkZd5qVadQrNWJsMs34VeFf-tD8gpsWJB5mJLx3Y732OrfLlQbD4z_WEjGrleygq7Nmv30JsgM1yewA-zibLOjnHiEXtfJe_Pv8ldNIlr286LxjTYtU3afBgqSqKnI7eusKCqfI7M1fx6M3CPPOOC2nTlYbpz8nv4goDANjUW6WT1CYTn11TKWgPTVMatNmkMRGMFdL99Wrc0jJkcVFByXyZbqpFQv8n-Ny6oPdOVH8OunQK5dA3iPsZC9uvRzAnN1PNdUOPAPg55mrAozcwCAck5VwBN0XDTGe4_ZcqpNYJCXn2h693OyVYvTeo3p1B2C8pCV6KclwFK0Op4_SHLldvhRajyQ/\" alt=\" \" style=\"display:none\"/><img src=\"https://us-east-sync.bidswitch.net/sync?ssp=openx&amp;dsp_id=25&amp;imp=1\" alt=\" \" style=\"display:none\"/>\n<iframe src='https://us-u.openx.net/w/1.0/pd?plm=5&ph=3d7a7809-a65a-442d-bf4e-f9613786ff87' width='0' height='0' style='display:none;'></iframe>  <div id='beacon_59859' style='position:absolute;left:0px;top:0px;visibility:hidden;'>\n    <img src='https://rtb-xv.openx.net/win/medianet?p=0.25&t=1fHJpZD1mYTYxMGZiMy1mODg3LTQ3YjMtYjY5ZS0xYWMyZWYzNWNhYWJ8cnQ9MTQ4MzM1NjQwMXxhdWlkPTUzODY1OTA3N3xhdW09RE1JRC5XRUJ8YXVwZj1kaXNwbGF5fHNpZD01MzcyOTkzNzV8cHViPTUzNzE0NTEwN3xwYz1VU0R8cmFpZD0zYTVjNjJmMi03ZDIwLTRiMGYtODU5Zi03MWQ5N2M2YjQzZmV8YWlkPTUzNzIyNzQ5NHx0PTEyfGFzPTMwMHgyNTB8bGlkPTUzNzE2NzE4OXxvaWQ9NTM3MDk2MTQ0fHA9MjEwfHByPTE0OXxhdGI9MjI4fGFkdj01MzcwNzMyNTZ8YWM9VVNEfHBtPVBSSUNJTkcuQ1BNfG09MXxhaT0xNTJkMDdiMS1kYTQxLTRjNDAtYWQxOC1mZTc0YzI2OTUzNGF8bWM9VVNEfG1yPTYxfHBpPTE0OXxtYT0yZjE0YzM2My03OWU5LTRiMDEtOWUzZS1iMTdjZTg0ZGQ0ZGR8bXJ0PTE0ODMzNTY0MDF8bXJjPVNSVF9XT058bXdhPTUzNzA3MzI1Nnxjaz0xfG13Ymk9Mzc1M3xtd2I9MjI4fG1hcD0yMTB8ZWxnPTF8bW9jPVVTRHxtb3I9NjF8bXBjPVVTRHxtcHI9MTQ5fG1wZj0xNDl8bW1mPTE0OXxtcG5mPTE0OXxtbW5mPTE0OXxwY3Y9MjAxNjEyMDV8bW89T1h8ZWM9MjVfMzYwNzI4fG1wdT0xNDl8bWNwPTIxMHxhcXQ9cnRifG13Yz01MzcwOTYxNDR8bXdwPTUzNzE2NzE4OXxtd2NyPTUzNzIyNzQ5NHxybm49MXxiYj0xfG13aXM9MXxtd3B0PW9wZW5ydGJfanNvbnx1cj1Kcndjdnl5MFBofGxkPWR5c29uLmNvbQ&c=USD&s=1'/>\n  </div>",	"publisher_id": "new-default",	"tp": "http://qsearch-a.akamaihd.net/log?logid=kfk&evtid=rtbstl&url=http%3A%2F%2Fhttp%253A%252F%252Fdigitach.net%252F&domain=http&size=300x250&ext_user_id=0&iid=0&auction_id=7b54575cd51c4286921bec5275301b54&ip=172.16.200.248&crid=183840587&b43b=10.0&bc=0.25&rtbW=43&app_dis=0&cc=RU&bname=&bid_type=-1&bmlevel=0<cm_win_macro>",	"size": "300x250",	"h": 250,	"w": 300,	"server_extras": {		"bid": 0.25,		"bidder_id": 43,		"og_bid": 10	},	"instl": 1,	"creative_type": "html",	"bidderInfoBean": {		"responseTime": 130	},	"view_width": 300,	"view_height": 250}]`)
	getRequest   = "GET /foobar?baz HTTP/1.1\r\nHost: google.com\r\nUser-Agent: aaa/bbb/ccc/ddd/eee Firefox Chrome MSIE Opera\r\n" +
		"Referer: http://xxx.com/aaa?bbb=ccc\r\nCookie: foo=bar; baz=baraz; aa=aakslsdweriwereowriewroire\r\n\r\n"
	postRequest = fmt.Sprintf("POST /foobar?baz HTTP/1.1\r\nHost: google.com\r\nContent-Type: foo/bar\r\nContent-Length: %d\r\n"+
		"User-Agent: Opera Chrome MSIE Firefox and other/1.2.34\r\nReferer: http://google.com/aaaa/bbb/ccc\r\n"+
		"Cookie: foo=bar; baz=baraz; aa=aakslsdweriwereowriewroire\r\n\r\n%s",
		len(fakeResponse), fakeResponse)
)

func benchmarkServerGet(b *testing.B, clientsCount, requestsPerConn int) {
	ch := make(chan struct{}, b.N)
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if !ctx.IsGet() {
				b.Fatalf("Unexpected request method: %s", ctx.Method())
			}
			ctx.Success("text/plain", fakeResponse)
			if requestsPerConn == 1 {
				ctx.SetConnectionClose()
			}
			registerServedRequest(b, ch)
		},
		Concurrency: 16 * clientsCount,
	}
	benchmarkServer(b, s, clientsCount, requestsPerConn, getRequest)
	verifyRequestsServed(b, ch)
}

func benchmarkNetHTTPServerGet(b *testing.B, clientsCount, requestsPerConn int) {
	ch := make(chan struct{}, b.N)
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method != "GET" {
				b.Fatalf("Unexpected request method: %s", req.Method)
			}
			h := w.Header()
			h.Set("Content-Type", "text/plain")
			if requestsPerConn == 1 {
				h.Set("Connection", "close")
			}
			w.Write(fakeResponse)
			registerServedRequest(b, ch)
		}),
	}
	benchmarkServer(b, s, clientsCount, requestsPerConn, getRequest)
	verifyRequestsServed(b, ch)
}

func benchmarkServerPost(b *testing.B, clientsCount, requestsPerConn int) {
	ch := make(chan struct{}, b.N)
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx){
			//t:=time.Now()
			resp := NewResponse()
			bidResps := &[]BidResponse{}
			
			BindFastHttp(ctx.PostBody(), bidResps)
			
			//if err != nil {
			//	fmt.Println(err)
			//}
			
			resp.Data.Add("ad", bidResps)
			
			resp.WriteFastHttp(ctx)
			//fmt.Println("time reqd is", time.Since(t))
			if requestsPerConn == 1 {
				ctx.SetConnectionClose()
			}
			registerServedRequest(b, ch)
		},
		//Handler: func(ctx *fasthttp.RequestCtx) {
		//	if !ctx.IsPost() {
		//		b.Fatalf("Unexpected request method: %s", ctx.Method())
		//	}
		//	body := ctx.Request.Body()
		//	if !bytes.Equal(body, fakeResponse) {
		//		b.Fatalf("Unexpected body %q. Expected %q", body, fakeResponse)
		//	}
		//	ctx.Success("text/plain", body)
		//	if requestsPerConn == 1 {
		//		ctx.SetConnectionClose()
		//	}
		//	registerServedRequest(b, ch)
		//},
		Concurrency: 16 * clientsCount,
	}
	benchmarkServer(b, s, clientsCount, requestsPerConn, postRequest)
	verifyRequestsServed(b, ch)
}

func benchmarkNetHTTPServerPost(b *testing.B, clientsCount, requestsPerConn int) {
	ch := make(chan struct{}, b.N)
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
			//t := time.Now()
			resp := NewResponse()
			bidResps := &[]BidResponse{}
			
			err := Bind(req.Body, bidResps)
			if err != nil {
				resp.Err = err
				return
			}
			
			resp.Data.Add("ad", bidResps)
			
			h := w.Header()
			if requestsPerConn == 1 {
				h.Set("Connection", "close")
			}
			resp.Write(w)
			registerServedRequest(b, ch)
			//fmt.Println("time is ", time.Since(t))
		}),
		//Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		//	if req.Method != "POST" {
		//		b.Fatalf("Unexpected request method: %s", req.Method)
		//	}
		//	body, err := ioutil.ReadAll(req.Body)
		//	if err != nil {
		//		b.Fatalf("Unexpected error: %s", err)
		//	}
		//	req.Body.Close()
		//	if !bytes.Equal(body, fakeResponse) {
		//		b.Fatalf("Unexpected body %q. Expected %q", body, fakeResponse)
		//	}
		//	h := w.Header()
		//	h.Set("Content-Type", "text/plain")
		//	if requestsPerConn == 1 {
		//		h.Set("Connection", "close")
		//	}
		//	w.Write(body)
		//	registerServedRequest(b, ch)
		//}),
	}
	benchmarkServer(b, s, clientsCount, requestsPerConn, postRequest)
	verifyRequestsServed(b, ch)
}

func registerServedRequest(b *testing.B, ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
		b.Fatalf("More than %d requests served", cap(ch))
	}
}

func verifyRequestsServed(b *testing.B, ch <-chan struct{}) {
	requestsServed := 0
	for len(ch) > 0 {
		<-ch
		requestsServed++
	}
	requestsSent := b.N
	for requestsServed < requestsSent {
		select {
		case <-ch:
			requestsServed++
		case <-time.After(100 * time.Millisecond):
			b.Fatalf("Unexpected number of requests served %d. Expected %d", requestsServed, requestsSent)
		}
	}
}

type realServer interface {
	Serve(ln net.Listener) error
}

func benchmarkServer(b *testing.B, s realServer, clientsCount, requestsPerConn int, request string) {
	ln := newFakeListener(b.N, clientsCount, requestsPerConn, request)
	ch := make(chan struct{})
	go func() {
		s.Serve(ln)
		ch <- struct{}{}
	}()
	
	<-ln.done
	
	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		b.Fatalf("Server.Serve() didn't stop")
	}
}
