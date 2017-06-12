package main

import (
	"github.com/rs/vast"
	//"github.com/golang/glog"
	"os"
	"io/ioutil"
	"encoding/xml"
	"fmt"
)

func loadFixture(path string) (*vast.VAST, error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()
	b, _ := ioutil.ReadAll(xmlFile)
	
	var v vast.VAST
	err = xml.Unmarshal(b, &v)
	return &v, err
}

func main() {
	v, err := loadFixture("testdata/media.xml")
	if err == nil {
		fmt.Println("first xml  ", v)
		fmt.Printf("%+v\n%+v\n", v, v.Ads[0].InLine)
	} else {
		fmt.Println("what man err!", err)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	v, err = loadFixture("testdata/google.xml")
	if err == nil {
		fmt.Println("second xml  ", v)
		fmt.Printf("%+v\n%+v\n", v, v.Ads[0].InLine)
	} else {
		fmt.Println("what man err!", err)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	v, err = loadFixture("testdata/brightroll.xml")
	if err == nil {
		fmt.Println("third xml  ", v)
		fmt.Printf("%+v\n%+v\n", v, v.Ads[0].InLine)
	} else {
		fmt.Println("what man err!", err)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	v, err = loadFixture("testdata/wrapper.xml")
	if err == nil {
		fmt.Println("third xml  ", v)
		fmt.Printf("%+v\n%+v\n", v, v.Ads[0].InLine)
	} else {
		fmt.Println("what man err!", err)
	}
}
