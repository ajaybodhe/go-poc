package main

import (
	"fmt"
	//"time"
	"github.com/media-net/cargo/db"
	//"github.com/jinzhu/gorm"
	//"encoding/json"
	//"github.com/lib/pq"
	//"strconv"

)

//type SlotSize struct {
//	ID      int
//	Size        string  `json:"size,omitempty"`
//	Epc        string  `json:"epc,omitempty"`
//}
//
//type BidderSlotsMap struct {
//	PublisherId     string `gorm:"type:text; primary_key" json:"-"`
//	ExternalPublisherId string `gorm:"-" json:"ecid,omitempty"`
//	ApSid string `gorm:"-" json:"sid,omitempty"`
//	CreativeId      string `gorm:"type:text; primary_key" json:"creative_id,omitempty"`
//	BidderId        int64  `gorm:"type:bigint; primary_key"  json:"bidder_id,omitempty"`
//	Sizes []SlotSize `gorm:"many2many:slot_sizes;" json:"sizes"`
//}

type SlotSize struct {
	Id               int64
	BidderSlotsMapId int64
	Size             string `json:"size,omitempty"`
	Epc              string `json:"epc,omitempty"`
}

type BidderSlotsMap struct {
	Id                  int64
	PublisherId         string     `json:"-"`
	ExternalPublisherId string     `sql:"-" json:"ecid,omitempty"`
	ApSid               string     `sql:"-" json:"sid,omitempty"`
	CreativeId          string     `json:"creative_id,omitempty"`
	BidderId            int64      `json:"bidder_id,omitempty"`
	Sizes               []SlotSize `json:"sizes"`
}

func main() {

	err := db.Connect("postgres://akshay@localhost:5432/doppelganger?sslmode=disable", 5, 10)
	if err != nil {
		fmt.Println(err)
	}
	db.Get().LogMode(true)

	fmt.Println(db.Get().AutoMigrate(&SlotSize{}))
	fmt.Println(db.Get().AutoMigrate(&BidderSlotsMap{}).AddUniqueIndex("unique_idx_pub_cr_bid", "publisher_id", "creative_id", "bidder_id"))

	bs := &BidderSlotsMap{}
	bs.PublisherId = "55"
	//bs.ExternalPublisherId = "40"
	bs.CreativeId = "50"
	//bs.ApSid = "abc"
	bs.BidderId = 123
	bs.Sizes = append(bs.Sizes, SlotSize{Size: "200x100", Epc: "1"})
	bs.Sizes = append(bs.Sizes, SlotSize{Size: "300x200", Epc: "2"})

	err = db.Get().Create(&bs).Error
	fmt.Println(err)

	bs = &BidderSlotsMap{}
	bs.PublisherId = "150"
	//bs.ExternalPublisherId = "40"
	bs.CreativeId = "159"
	bs.BidderId = 456
	bs.Sizes = []SlotSize{}
	bs.Sizes = append(bs.Sizes, SlotSize{Size: "400x300", Epc: "2"})
	bs.Sizes = append(bs.Sizes, SlotSize{Size: "300x500", Epc: "2"})
	err = db.Get().Create(&bs).Error
	fmt.Println(err)

	// DELETE AND CREATE AFRESH
	//err = db.Get().Model(&bs).Related(&slots).Error
	//fmt.Println(err)
	//err = db.Get().Delete(&bs).Error
	//fmt.Println(err)
	//err = db.Get().Delete(&slots).Error
	//fmt.Println(err)

	bs1 := &BidderSlotsMap{}
	bs1.PublisherId = "55"
	bs1.CreativeId = "50"
	bs1.BidderId = 123
	err = db.Get().Where("publisher_id = ? and creative_id = ? and bidder_id = ?",
		bs1.PublisherId, bs1.CreativeId, bs1.BidderId).Find(&bs1).Error
	err = db.Get().Model(*bs1).Where("size = ?", "300x200").Related(&bs1.Sizes).Error
	fmt.Println("data issssss, ", err, bs1)

	//bs.YbncaAutoShare = 77
	//err = db.Get().Save(&bs).Error
	//fmt.Println(err)
}
