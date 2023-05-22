package luckrand

import (
	"fmt"
	"testing"
)

func BenchmarkB(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = RandomString(10)
	}
}

func TestRand(t *testing.T) {
	str := RandomString(10)
	fmt.Println(str)
}

func TestIdMakerP(t *testing.T) {
	idm := IdMakerInOnP{}
	idm.SetStartId(1000)
	fmt.Println(idm.Get())
	fmt.Println(idm.Get())
	fmt.Println(idm.Get())
	fmt.Println(idm.Get())
	idm.SetStartId(1000)
	fmt.Println(idm.Get())
	fmt.Println(idm.Get())
	fmt.Println(idm.Get())

	counter := Counter{}
	go func() {
		fmt.Println(counter.Add(), "end")
	}()
	fmt.Println(counter.Get())
}
