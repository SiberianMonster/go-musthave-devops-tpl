package main

import (

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metricsfuncs"
	"fmt"
)




func main() {
	err := metricsfuncs.ReportUpdate(2,10)
	if err != nil {fmt.Println(err)
		}

}