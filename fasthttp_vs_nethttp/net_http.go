package main

import (
	"encoding/json"
	"time"
	"net/http"
	"fmt"
	//"errors"
	"io"
)


func Bind(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	err := json.NewDecoder(body).Decode(v)
	//logger.D("Bind data: ", fmt.Sprintf("%v", v))
	return err
}

func TestHello(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	resp := NewResponse()
	bidResps := &[]BidResponse{}
	
	err := Bind(r.Body, bidResps)
	if err != nil {
		resp.Err = err
		return
	}
	
	resp.Data.Add("ad", bidResps)
	resp.Write(w)
	fmt.Println("time is ", time.Since(t))
}
//
//func main() {
//
//}
