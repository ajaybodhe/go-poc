package main

import (
	"fmt"
	"encoding/json"
	"github.com/lib/pq"dd
	"database/sql"
	"github.com/jinzhu/gorm"
	//"strconv"
)

type SupportedAds struct {
	TagId       string `gorm:"type:text; primary_key" json:"tag_id"`
	PublisherId string `gorm:"type:text; primary_key" json:"publisher_id"`
	Supported   pq.Int64Array `gorm:"type:integer[]" json:"supported"`
	Sstr pq.StringArray `gorm:"type:text[]" json:"sstr"`
}

type Postgres struct {
	db *gorm.DB
}

func Connect(url string, maxIdleConnections, maxOpenConnections int) (*Postgres, error) {
	var err error
	pg := new(Postgres)
	pg.db, err = gorm.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	// Get database connection handle [*sql.DB](http://golang.org/pkg/database/sql/#DB)
	pg.db.DB()
	// Then you could invoke `*sql.DB`'s functions with it
	err = pg.db.DB().Ping()
	if err != nil {
		return nil, err
	}
	pg.db.SetLogger(PgLogger{})
	pg.db.LogMode(true)
	pg.db.DB().SetMaxIdleConns(maxIdleConnections)
	pg.db.DB().SetMaxOpenConns(maxOpenConnections)
	pg.db.SingularTable(false)
	return pg, nil
}
}
func main() {
	err := db.Connect("postgres://akshay@localhost:5432/doppelganger?sslmode=disable", 5, 10)
	if err != nil {
		fmt.Println(err)
	}
	
	fmt.Println(db.CreateTable(&SupportedAds{}))
	
	s := SupportedAds{TagId:"12345",
				PublisherId:"ABCDE",
				Supported: pq.Int64Array{1,2,3},
				Sstr:pq.StringArray{"video","audio"},}
	fmt.Println(db.Create(&s))
	
	var s1 SupportedAds
	fmt.Println(db.Get().Where("tag_id = ? and publisher_id = ?", s.TagId, s.PublisherId).Find(&s1))
	
	fmt.Println(s1)
	sj,_ := json.Marshal(s1)
	fmt.Println(string(sj))
}
