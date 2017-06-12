package main

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"unsafe"
	"time"
)

const (
	SomeThingStructPool  = "SomeThingStructPool"
	ManyThingsStructPool = "ManyThingsStructPool"
	PoolElementsCount    = 24
	Uint8BitCOunt        = 8
)

type Something struct {
	Name string
}

func (s *Something) Init() {
	fmt.Println("inside init() of", s)
	s.Name = ""
}

func (s *Something) Destroy() {
	fmt.Println("inside destroy() of", s)
	s.Name = ""
}

type ManyThings struct {
	Name            string
	Id              string
	SomethingSlices []*Something
	SomethingMap    map[string]*Something
}

func (m *ManyThings) Init() {
	fmt.Println("inside init() of", m)
	m.SomethingMap = make(map[string]*Something)
}

/*
	Releasing children struct object here
*/
func (m *ManyThings) Destroy() {
	fmt.Println("inside destroy() of", m)

	for i := range m.SomethingSlices {
		somethingPool.PutPool(m.SomethingSlices[i])
	}
	for i := range m.SomethingMap {
		somethingPool.PutPool(m.SomethingMap[i])
	}
}

// This will hold a pool for a given structure
// freeList: bitarray having one bit per block
// currFreeListIndex: current byte with free block
// mutex: synchronize pool operations
// actualpool: slice of actual struct pool, return one object addr from it
// sizeOfStruct: size of the stuct, use unsafe.Sizeof(struct)
type MemPool struct {
	name              string
	freeList          []uint8
	currFreeListIndex int
	mutex             *sync.Mutex
	actualPool        reflect.Value
	sizeOfStruct      uintptr
}

// every struct has to implement this interface before creating a pool
// init: initialise elements, say for a map call make()
// destroy: set all elements to nil value sosay if a map was allocated , it will go for gc
// TBD handle init, destroy automatically thorugh reflection
type PoolInterface interface {
	Init()
	Destroy()
}

/*
	input : slice of struct for which pool is to be created
	return: pointer to new MemPool and error if any
*/
func CreatePool(name string, slc interface{}, sizeofstruct uintptr) (*MemPool, error) {

	pl := reflect.ValueOf(slc)
	if math.Mod(float64(pl.Len()), float64(Uint8BitCOunt)) != 0 {
		return nil, errors.New("Slice size has to be multiple of 8")
	}
	if _, ok := pl.Index(0).Addr().Interface().(PoolInterface); !ok {
		return nil, errors.New("Structure has to implement PoolInterface")
	}
	pool := MemPool{
		// to handle pool of 24 structs,
		// so u need one bit per struct that is 3 uint8 to manage that
		freeList:          make([]uint8, pl.Len()/Uint8BitCOunt),
		currFreeListIndex: 0,
		mutex:             &sync.Mutex{},
		actualPool:        pl,
		sizeOfStruct:      sizeofstruct,
		name:              name,
	}

	return &pool, nil
}

// returns the available block anf bit position to be freed up later
func (pool *MemPool) GetPool() (*reflect.Value, error) {

	var pos int
	pos = -1

	pool.mutex.Lock()
	
	t:= time.Now()
	
	currIndex := pool.currFreeListIndex
	for pos < 0 {
		if pool.freeList[pool.currFreeListIndex]&0xff == 0xff {
			// all bits set get the next location
			pool.currFreeListIndex = int(math.Mod(float64(pool.currFreeListIndex+1), float64(len(pool.freeList))))
			//fmt.Println(pool.name, " pool.currFreeListIndex is", pool.currFreeListIndex)
			if pool.currFreeListIndex == currIndex {
				pool.mutex.Unlock()
				return nil, errors.New(pool.name + " No free block !")
			}
		} else {
			// get the first rightmost zero bit
			pos = int((^pool.freeList[pool.currFreeListIndex]) & (pool.freeList[pool.currFreeListIndex] + 1))
		}
	}
	pool.freeList[pool.currFreeListIndex] = pool.freeList[pool.currFreeListIndex] | uint8(pos)
	//fmt.Println(pool.name, " pos is:", pos, pool.freeList[pool.currFreeListIndex])
	pos = int(math.Log2(float64(pos)))
	pos = pos + pool.currFreeListIndex*Uint8BitCOunt
	fmt.Println(pool.name, " POOLGET: current free list index:", pool.currFreeListIndex, " free pos:", pos)
	
	v := pool.actualPool.Index(pos).Addr()
	fmt.Println(pool.name, " time for getpool ", time.Since(t))
	
	pool.mutex.Unlock()
	
	// call init on struct
	v.Interface().(PoolInterface).Init()

	return &v, nil
}

func (pool *MemPool) PutPool(ele interface{}) error {
	// call destroy on struct
	ele.(PoolInterface).Destroy()
	
	pool.mutex.Lock()
	
	t:= time.Now()
	
	pos := (reflect.ValueOf(ele).Pointer() - pool.actualPool.Index(0).Addr().Pointer()) / pool.sizeOfStruct
	currIndex := pos / Uint8BitCOunt
	pos1 := uint(math.Mod(float64(pos), float64(Uint8BitCOunt)))
	pool.freeList[currIndex] &= (^(1 << pos1))
	
	fmt.Println(pool.name, " POOLPUT: current free list index:", currIndex, " free pos:", pos1, " passed pos", pos)
	
	fmt.Println(pool.name, " time for putpool ", time.Since(t))
	
	pool.mutex.Unlock()
	
	return nil
}

/*
These pools have to be global
*/
var manythingPool *MemPool
var somethingPool *MemPool

func main() {
	var err error

	manyThingSlice := make([]ManyThings, PoolElementsCount)
	manythingPool, err = CreatePool(ManyThingsStructPool, manyThingSlice, unsafe.Sizeof(ManyThings{}))
	if err != nil {
		fmt.Println(err)
	}

	someThingSlice := make([]Something, PoolElementsCount)
	somethingPool, err = CreatePool(SomeThingStructPool, someThingSlice, unsafe.Sizeof(Something{}))
	if err != nil {
		fmt.Println(err)
	}

	mi, _ := manythingPool.GetPool()
	m := mi.Interface().(*ManyThings)

	si1, _ := somethingPool.GetPool()
	si2, _ := somethingPool.GetPool()
	si3, _ := somethingPool.GetPool()
	si4, _ := somethingPool.GetPool()

	s1 := si1.Interface().(*Something)
	s2 := si2.Interface().(*Something)
	s3 := si3.Interface().(*Something)
	s4 := si4.Interface().(*Something)

	m.SomethingMap["s1"] = s1
	m.SomethingMap["s2"] = s2

	m.SomethingSlices = append(m.SomethingSlices, s3)
	m.SomethingSlices = append(m.SomethingSlices, s4)

	manythingPool.PutPool(m)

}
