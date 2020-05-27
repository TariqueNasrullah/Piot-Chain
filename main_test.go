package main

import "testing"

func TestAdd(t *testing.T)  {
	res := Add(10, 20)
	if res != 30 {
		t.Fatal("Wrong Result")
	}
}