package main

import (
	"github.com/Workiva/go-datastructures/trie/yfast"
	"fmt"
	//"encoding/binary"
	//"bytes"
	"unsafe"
)
type GeoDetails struct {
	City string
	Region string
	Zip string
	Country string
	Lat float64
	Lon float64
	Tz int
}
type IpToGeoDetails struct {
	Ip uint64
	//GeoDet *GeoDetails
	City string
	Region string
	Zip string
	Country string
	Lat float64
	Lon float64
	Tz int32
}

func(i *IpToGeoDetails) Key() uint64 {
	return i.Ip
}

func main() {
	yt := yfast.New(uint64)
	
	for i:=0; i< 300000; i=i+2 {
		ip := new(IpToGeoDetails)
		ip.Ip = uint64(i)
		yt.Insert(ip)
	}
	
	//buf := new(bytes.Buffer)
	//err := binary.Write(buf, binary.LittleEndian, yt)
	//if err != nil {
	//	fmt.Println("Error in binary : ", err)
	//	return
	//}
	fmt.Println("Binary size ", unsafe.Sizeof(IpToGeoDetails{}))
	for i:=2; i< 100; i=i+2 {
		ip := yt.Get(uint64(i))
		//fmt.Println("ip is ", i, " and retrived ", ip.(*IpToGeoDetails).Ip)
		ipO := yt.Get(uint64(i+1))
		//fmt.Println("ip is ", i +1 , " and retrived ", ipO == nil)
		ipP := yt.Predecessor(uint64(i+1))
		//fmt.Println("ip is ", i+1, " and retrived is ", ipP.(*IpToGeoDetails).Ip)
		
		//fmt.Println("\n\n\n\n\n ")
	}
	
}