package main

import (
	"fmt"
	"time"

	"github.com/TaceyWong/go-aps/aps"
)

func task(id int) {
	fmt.Printf("I am runnning task.#%d\n",id)
}



func main() {
	var sigList []chan bool
	s1 := aps.NewScheduler()
	s1.Every(1).Seconds().Do(task,1)
	sigList = append(sigList, s1.Start())
	s2 := aps.NewScheduler()
	s2.Every(1).Seconds().Do(task,2)
	sigList = append(sigList, s2.Start())
	
	s3 := aps.NewScheduler()
	s3.Every(1).Seconds().Do(task,3)
	sigList = append(sigList, s3.Start())

	s4 := aps.NewScheduler()
	s4.Every(1).Seconds().Do(task,4)
	sigList = append(sigList, s4.Start())

	for index,sig:= range sigList{
		time.Sleep(time.Second*10)
		fmt.Printf("To stop the %d(th) scheduler\n",index+1)
		sig <- true
	}
	

}
