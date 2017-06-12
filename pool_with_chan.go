package main

import (
	"errors"
	"fmt"
)

const (
	// pool names
	somethingPool = "somethingPool"
	manythingPool = "manythingPool"

	// no of elements in pool
	PoolElementsCount = 2
)

// Lets define simple struct for pooling
type Something struct {
	Name string
}

/*
 Init()
 is necessary to set the values of this struct block to default ones,
 since the same block is being used again n again
*/
func (s *Something) Init() {
	s.Name = ""
}

/*
 Destroy()
 is necessary to set the allocated variables of this struct block to nil,
 these allocated variables wil then be open for Garbage COllection
*/
func (s *Something) Destroy() {
	// TODO basic struct, nothing to destroy
}

/*
  NewSomething()
  is necessary if the pool channel currently has no block to return.
  Pool will use NewSomething() to create ne one & return the same.
*/
func NewSomething() PoolInterface {
	return &Something{}
}

// Lets define another struct but a bit complex
type Manything struct {
	Name            string
	Id              int
	SomethingSlices []*Something
	SomethingMap    map[string]*Something
}

func (m *Manything) Init() {
	m.Name = ""
	m.Id = 0
	m.SomethingSlices = make([]*Something, 0)
	m.SomethingMap = make(map[string]*Something)
}

func (m *Manything) Destroy() {
	m.SomethingMap = nil
	m.SomethingSlices = nil
}

func NewManything() PoolInterface {
	return &Manything{}
}

/*
 every struct has to implement this interface before creating a pool
 init: initialise elements, say for a creating member map variable  call make()
  destroy: set all elements to nil value, so say if a map was allocated , it will go for gc
*/
// TODO handle init, destroy automatically thorugh reflection
type PoolInterface interface {
	Init()
	Destroy()
}

// function to create new struct block if existing ones are exhausted
type New func() PoolInterface

// This will hold a pool for a given structure
type myPool struct {
	data chan PoolInterface
	New
	//mutex *sync.Mutex
}

// keep map of all pools by struct name as key
var poolMap map[string]*myPool

func init() {
	poolMap = make(map[string]*myPool)
}

// create a ne pool for ne struct
func CreatePool(name string, n New) error {
	if _, ok := poolMap[name]; ok {
		return errors.New("Pool Already Exists")
	}
	pool := myPool{
		data: make(chan PoolInterface, PoolElementsCount),
		New:  n,
	}
	poolMap[name] = &pool
	return nil
}

// returns the available block
func GetPool(name string) (PoolInterface, error) {

	var d PoolInterface

	if pool, ok := poolMap[name]; ok {

		select {
		case d = <-pool.data:
		default:
			d = pool.New()
		}

		d.Init()
		return d, nil
	}
	return nil, errors.New("No Pool by this name found")
}

// when the use of a struct block is done
func PutPool(name string, d PoolInterface) error {
	if pool, ok := poolMap[name]; ok {
		d.Destroy()
		pool.data <- d
		return nil
	}
	return errors.New("No Pool by this name found")
}

func main() {
	///////////////////////////////// POOL 1

	err := CreatePool(somethingPool, NewSomething)
	if err != nil {
		fmt.Println(err)
	}

	s, err := GetPool(somethingPool)
	sp := s.(*Something)
	if err != nil {
		fmt.Println(err)
	}

	sp.Name = "Ajay"
	err = PutPool(somethingPool, sp)
	if err != nil {
		fmt.Println(err)
	}

	s, err = GetPool(somethingPool)
	sp = s.(*Something)
	if err != nil {
		fmt.Println(err)
	}
	// THIS WILL NOT PRINT AJAY ONLY
	fmt.Println("sp.Name is ", sp.Name)
	///////////////////////////////// POOL 1

	///////////////////////////////// POOL 2
	err = CreatePool(manythingPool, NewManything)
	if err != nil {
		fmt.Println(err)
	}

	m, err := GetPool(manythingPool)
	mp := m.(*Manything)
	if err != nil {
		fmt.Println(err)
	}

	mp.Name = "Ajay"
	mp.Id = 123
	mp.SomethingSlices = append(mp.SomethingSlices, sp)
	mp.SomethingMap = make(map[string]*Something)
	mp.SomethingMap["mykey"] = sp
	err = PutPool(manythingPool, mp)
	if err != nil {
		fmt.Println(err)
	}

	m, err = GetPool(manythingPool)
	mp = m.(*Manything)
	if err != nil {
		fmt.Println(err)
	}
	// THIS WILL NOT PRINT AJAY, ABCD123 ONLY
	fmt.Println("sp.Name is ", mp.Name, mp.Id)
	///////////////////////////////// POOL 2
}
