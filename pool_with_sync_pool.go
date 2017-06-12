package main

import "sync"
import (
	"errors"
	"fmt"
)

const (
	somethingPool  = "somethingPool"
	somethingPool1 = "somethingPool1"
)

type Something struct {
	Name string
}

func PoolRegisterSomething() interface{} {
	return &Something{}
}

type Something1 struct {
	Name string
	Id   string
}

func PoolRegisterSomething1() interface{} {
	return &Something1{}
}

type PoolRegister func() interface{}

var poolMap map[string]*sync.Pool

func createPool(name string, pr PoolRegister) error {
	if _, ok := poolMap[name]; ok {
		return errors.New("Pool Already Exists")
	}
	pool := sync.Pool{
		New: pr,
	}
	poolMap[name] = &pool
	return nil
}

func GetPool(name string) (interface{}, error) {

	if pool, ok := poolMap[name]; ok {
		return pool.Get(), nil
	}
	return nil, errors.New("No Pool by this name found")
}

func PutPool(name string, x interface{}) error {
	if pool, ok := poolMap[name]; ok {
		pool.Put(x)
		return nil
	}
	return errors.New("No Pool by this name found")
}

func init() {
	poolMap = make(map[string]*sync.Pool)
}

func main() {
	///////////////////////////////// POOL 1
	err := createPool(somethingPool, PoolRegisterSomething)
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
	// THIS WILL PRINT AJAY ONLY
	fmt.Println("sp.Name is ", sp.Name)
	///////////////////////////////// POOL 1

	///////////////////////////////// POOL 2
	err = createPool(somethingPool1, PoolRegisterSomething1)
	if err != nil {
		fmt.Println(err)
	}

	s1, err := GetPool(somethingPool1)
	sp1 := s1.(*Something1)
	if err != nil {
		fmt.Println(err)
	}

	sp1.Name = "Ajay"
	sp1.Id = "ABCD123"
	err = PutPool(somethingPool1, sp1)
	if err != nil {
		fmt.Println(err)
	}

	s1, err = GetPool(somethingPool1)
	sp1 = s1.(*Something1)
	if err != nil {
		fmt.Println(err)
	}
	// THIS WILL PRINT AJAY, ABCD123 ONLY
	fmt.Println("sp.Name is ", sp1.Name, sp1.Id)
	///////////////////////////////// POOL 2
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8000", nil)
}
