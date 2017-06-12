package main

import (
	"github.com/valyala/fasthttp"
	"encoding/json"
	"errors"
	"net/http"
	//"time"
	//"fmt"
)

var (
	DefaultErrorCode = fasthttp.StatusConflict
)
type JSONResponse map[string]interface{}

func (r JSONResponse) String() (res string) {
	json, err := json.Marshal(r)
	if err != nil {
		res = ""
		return
	}
	res = string(json)
	return
}

func (r JSONResponse) Add(key string, value interface{}) {
	(map[string]interface{}(r))[key] = value
}

func (r JSONResponse) ByteArray() (res []byte) {
	json, err := json.Marshal(r)
	if err != nil {
		res = nil
		return
	}
	res = json
	return
}

type Response struct {
	Msg     string
	Data    JSONResponse
	Status  int
	Err     error
	Written bool
	
	Log string
}

func NewResponse() Response {
	r := Response{}
	r.Data = JSONResponse{}
	r.Status = -1
	r.Msg = ""
	r.Written = false
	return r
}

func (r *Response) WriteFastHttp(ctx *fasthttp.RequestCtx) {
	if r.Written {
		return
	}
	if r.Err != nil || r.Status > 399 {
		r.writeErrorResponseFastHttp(ctx)
		return
	}
	if r.Status == fasthttp.StatusFound {
		// redirect do not write anything
		return
	}
	r.writeResponseFastHttp(ctx)
}

func (r *Response) writeResponseFastHttp(ctx *fasthttp.RequestCtx) {
	
	ctx.Response.Header.Add("Content-Type", "application/json")
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Auth-Key, Session-Key")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	if r.Status == -1 {
		r.Status = fasthttp.StatusOK
	}
	ctx.SetStatusCode(r.Status)
	res := JSONResponse{
		"success":     true,
		"data":        r.Data,
		"message":     r.Msg,
		"api_version": 1,
	}
	ctx.Write(res.ByteArray())
}

func (r *Response) writeErrorResponseFastHttp(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Add("Content-Type", "application/json")
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Auth-Key, Sdk-Version, OPTIONS")
	if r.Status == -1 {
		ctx.SetStatusCode(DefaultErrorCode)
	} else {
		ctx.SetStatusCode(r.Status)
	}
	if r.Err == nil {
		switch r.Status {
		case fasthttp.StatusUnauthorized:
			r.Err = errors.New("Unauthorized access")
		default:
			r.Err = errors.New("Illegal request")
		}
		
	}
	res := JSONResponse{
		"errors":        []string{r.Err.Error()},
		"error_message": r.Msg,
		"success":       false,
	}
	ctx.Write(res.ByteArray())
}

func (r *Response) Write(w http.ResponseWriter) {
	if r.Written {
		return
	}
	if r.Err != nil || r.Status > 399 {
		r.writeErrorResponse(w)
		return
	}
	if r.Status == http.StatusFound {
		// redirect do not write anything
		return
	}
	r.writeResponse(w)
}

func (r *Response) writeResponse(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Auth-Key, Session-Key")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	if r.Status == -1 {
		r.Status = http.StatusOK
	}
	w.WriteHeader(r.Status)
	res := JSONResponse{
		"success":     true,
		"data":        r.Data,
		"message":     r.Msg,
		"api_version": 1,
	}
	w.Write(res.ByteArray())
}

func (r *Response) writeErrorResponse(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Auth-Key, Sdk-Version, OPTIONS")
	if r.Status == -1 {
		w.WriteHeader(DefaultErrorCode)
	} else {
		w.WriteHeader(r.Status)
	}
	if r.Err == nil {
		switch r.Status {
		case http.StatusUnauthorized:
			r.Err = errors.New("Unauthorized access")
		default:
			r.Err = errors.New("Illegal request")
		}
		
	}
	res := JSONResponse{
		"errors":        []string{r.Err.Error()},
		"error_message": r.Msg,
		"success":       false,
	}
	w.Write(res.ByteArray())
}

type SeatBid struct {
	Bid   []Bid  `json:"bid,omitempty"`
	Seat  string `json:"seat,omitempty"`
	Group int64  `json:"group,omitempty"`
	Ext   Ext    `json:"ext,omitempty"`
}

type Bid struct {
	Id             string   `json:"id"`
	ImpId          string   `json:"impid"`
	Price          float64  `json:"price"`
	NUrl           string   `json:"nurl,omitempty"`
	BUrl           string   `json:"burl,omitempty"`
	LUrl           string   `json:"lurl,omitempty"`
	ADM            string   `json:"adm,omitempty"`
	AdId           string   `json:"adid,omitempty"`
	ADomain        []string `json:"adomain,omitempty"`
	Bundle         string   `json:"bundle,omitempty"`
	IUrl           string   `json:"iurl,omitempty"`
	CId            string   `json:"cid,omitempty"`
	CrId           string   `json:"crid,omitempty"`
	Tactic         string   `json:"tactic,omitempty"`
	Cat            []string `json:"cat,omitempty"`
	Attr           []int64  `json:"attr,omitempty"`
	Api            int64    `json:"api,omitempty"`
	Protocol       int64    `json:"protocol,omitempty"`
	QAGMediaRating int64    `json:"qagmediarating,omitempty"`
	Language       string   `json:"language,omitempty"`
	DealId         string   `json:"dealid,omitempty"`
	Width          int64    `json:"w,omitempty"`
	Height         int64    `json:"h,omitempty"`
	WRatio         int64    `json:"wratio,omitempty"`
	HRatio         int64    `json:"hratio,omitempty"`
	Exp            int64    `json:"exp,omitempty"`
	Ext            Ext      `json:"ext,omitempty"`
}

type BidderInfoBean struct {
	ResponseTime     int      `json:"responseTime,omitempty"`
	Category         []string `json:"category,omitempty"`
	CampaignId       string   `json:"cmpId,omitempty"`
	DataCenter       string   `json:"dataCenter,omitempty"`
	ServerId         string   `json:"serverId,omitempty"`
	BuyerMemberId    string   `json:"buyerMemberId,omitempty"`
	BrandId          string   `json:"brandId,omitempty"`
	ProviderBidderId string   `json:"providerBidderId,omitempty"`
	AdvBrandName     string   `json:"advBrandName,omitempty"`
	AdvUrl           string   `json:"advUrl,omitempty"`
	DealId           string   `json:"di,omitempty"`
	DealType         string   `json:"dt,omitempty"`
	AdType           string   `json:"adtype,omitempty"`
	//ProviderRequestId string   `json:"prvReqId,omitempty"`
	//ProviderAccountId string   `json:"prvAccId,omitempty"`
	//ProviderApiId     string   `json:"prvApiId,omitempty"`
	//ProviderName      string   `json:"pvnm,omitempty"`
}

// BidResponse Model
type BidResponse struct {
													   // this is not strict open rtb response,
													   // so Id field here corresponds to Id field in BidLandscape
	Id         string    `json:"id,omitempty"`
	SeatBid    []SeatBid `json:"seatbid,omitempty"`
	BidId      string    `json:"bidid,omitempty"`
	Currency   string    `json:"cur,omitempty"`
	CustomData string    `json:"customdata,omitempty"`
	NBR        int64     `json:"nbc,omitempty"`
	Ext        Ext       `json:"ext,omitempty"`
	
	BidderId      int64          `json:"bidder_id,omitempty"`
	CreativeId    string         `json:"creative_id,omitempty"`
	AdType        string         `json:"adtype,omitempty"`
	AdCode        string         `json:"adcode,,omitempty"`
	OgBid         float32        `json:"og_bid,omitempty"`
	Bid           float32        `json:"bid,omitempty"`
	PublisherId   string         `json:"publisher_id,,omitempty"`
	Variant       int            `json:"variant,omitempty"`
	TrackingPixel string         `json:"tp,omitempty"`
	Size          string         `json:"size,omitempty"`
	Height        int            `json:"h,omitempty"`
	Width         int            `json:"w,omitempty"`
	LoggingPixels []string       `json:"logging_pixels,omitempty"`
	ServerExtras  Ext            `json:"server_extras,omitempty"`
	Interstitial  int8           `json:"instl,omitempty"`
	CreativeType  string         `json:"creative_type"`
	Bl            []BidLandscape `json:"bl,omitempty"` // Bid Landscape
	Bib           BidderInfoBean `json:"bidderInfoBean,omitempty"`
	AuctionWinUrl string         `json:"auction_win_url,omitempty"`
	ViewWidth     int            `json:"view_width,omitempty"`
	ViewHeight    int            `json:"view_height,omitempty"`
}

type BidLandscape struct {
	//Id             string         `json:"id,omitempty"`
	Bib              BidderInfoBean `json:"bidderInfoBean,omitempty"`
	Uid              string         `json:"uid,omitempty"`
	NoBid            bool           `json:"no_bid,omitempty"`
	NoBidReason      int            `json:"nbc,omitempty"`
	BidderId         int64          `json:"bidder_id,omitempty"`
	Fb               bool           `json:"fb,omitempty"`
	AdCode           string         `json:"adcode,omitempty"`
	LoggingPixels    []string       `json:"logging_pixels,omitempty"`
	OgBid            float32        `json:"og_bid,omitempty"`
	Bid              float32        `json:"bid,omitempty"`
	CustomerBidPrice float32        `json:"cbdp,omitempty"`
	Size             string         `json:"size,omitempty"`
	Status           int            `json:"s,omitempty"`
	StatusDesc       string         `json:"snm,omitempty"`
	//SubBidderId    int64          `json:"sbdrid,omitempty"`
	//BaseBidderId   int64          `json:"bbdrid,omitempty"`
	// TBD Bid may or may not come and need to be calculated in auction
	//ProviderDiscrepencyShare float32 `json:"adj1,omitempty"`
	//ProviderRevenueShare     float32 `json:"adj2,omitempty"`
	//ClosePrice               float32        `json:"clsPrc,omitempty"`
	//DfpBid                   float32        `json:"dfpbd,omitempty"`
	//BidderInfo               string         `json:"binfobid,omitempty"`
}
type Ext map[string]interface{}


