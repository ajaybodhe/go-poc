package main

import (
	"fmt"
)

type DQNode struct {
	Key, Data  int
	Next, Prev *DQNode
}
type LRUCache struct {
	Head    *DQNode
	NodeMap map[int]*DQNode
	Size    int
}

func NewLRUCache(size int) *LRUCache {
	L := new(LRUCache)
	L.Size = size
	L.NodeMap = make(map[int]*DQNode)
	L.Head = nil
	return L
}
func (L *LRUCache) Print() {
	if L.Head == nil {
		fmt.Print("No List")
		return
	}
	fmt.Println("Map is :")
	for k, v := range L.NodeMap {
		fmt.Println(k,":",v.Data)
	}
	
	fmt.Print("List is : ", L.Head.Data)
	for node := L.Head.Next; node != L.Head; node = node.Next {
		fmt.Print(", ", node.Data)
	}
	fmt.Println()	
	fmt.Println()
}
func (L *LRUCache) Get(key int) (int, bool) {
	if node, ok := L.NodeMap[key]; ok {
		if len(L.NodeMap) == 1 || node == L.Head {								
			return node.Data, true
		}
		node.Prev.Next = node.Next
		node.Next.Prev = node.Prev
		delete(L.NodeMap, node.Key)
		
		node.Next = L.Head
		node.Prev = L.Head.Prev
		L.Head.Prev.Next = node
		L.Head.Prev = node
		L.Head = node
		L.NodeMap[key] = L.Head
		
		return L.Head.Data, true
	} 
	return -1, false
}
func (L *LRUCache) Insert(key, val int) {
	if L.Head == nil {
		L.Head = new(DQNode)
		L.Head.Data = val
		L.Head.Key = key
		L.Head.Next = L.Head
		L.Head.Prev = L.Head
		L.NodeMap[key] = L.Head
		return
	}
	var node *DQNode
	var ok bool
	if node, ok = L.NodeMap[key]; ok {
		if len(L.NodeMap) == 1 {
			L.Head.Data = val						
			return
		}
		node.Prev.Next = node.Next
		node.Next.Prev = node.Prev
		delete(L.NodeMap, node.Key)
	} else {
		if len(L.NodeMap) == L.Size {
			if L.Size == 1 {
				L.Head.Data = val
				L.Head.Key = key
				return
			}
			// first delete keys
			delnode := L.Head.Prev
			delnode.Next.Prev = delnode.Prev
			delnode.Prev.Next = delnode.Next
			delete(L.NodeMap, delnode.Key)
		}
		// create new node
		node = new(DQNode)
	}
	node.Key = key
	node.Data = val
	node.Next = L.Head
	node.Prev = L.Head.Prev
	L.Head.Prev.Next = node
	L.Head.Prev = node
	L.Head = node
	L.NodeMap[key] = L.Head
}
func main() {
	// insert new keys
	L := NewLRUCache(5)
	L.Insert(12, 144)
	L.Print()
	L.Insert(11, 121)
	L.Print()
	L.Insert(10, 100)
	L.Print()
	L.Insert(9, 81)
	L.Print()
	L.Insert(8, 64)
	L.Print()
	
	// enter new keys, displacing older ones
	L.Insert(7, 49)
	L.Print()
	L.Insert(6, 36)
	L.Print()

	// test cases: modify existing keys	
	L.Insert(7, 70)
	L.Print()
	L.Insert(9, 90)
	L.Print()

	// get existing keys
	val, ok := L.Get(8)
	fmt.Println("Key search status :", val, ok)
	L.Print()
	val, ok = L.Get(10)
	fmt.Println("Key search status :", val, ok)
	L.Print()
	
	// find non existing keys
	val, ok = L.Get(11)
	fmt.Println("Key search status :", val, ok)
	//L.Print()
}
