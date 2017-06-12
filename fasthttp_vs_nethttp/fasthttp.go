package main

import (
	"encoding/json"
	"time"
	"github.com/valyala/fasthttp"
	"fmt"
	//"errors"
	"net/http"
)
func BindFastHttp(body []byte, v interface{}) error {
	//defer body.Close()
	err := json.Unmarshal(body, v)
	//logger.D("Bind data: ", fmt.Sprintf("%v", v))
	return err
}
func TestHelloFastHttp(ctx *fasthttp.RequestCtx) error {
	t:=time.Now()
	resp := NewResponse()
	bidResps := &[]BidResponse{}
	
	err := BindFastHttp(ctx.PostBody(), bidResps)
	if err != nil {
		resp.Err = err
		return nil
	}
	
	resp.Data.Add("ad", bidResps)
	
	resp.WriteFastHttp(ctx)
	fmt.Println("time reqd is", time.Since(t))
	return nil
}

func main() {
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/api/v1/hello/test":
			TestHelloFastHttp(ctx)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}
	fasthttp.ListenAndServe(":9000", m)
	
	http.HandleFunc("/api/v1/hello/test", TestHello)
	http.ListenAndServe(":9001", nil)
}