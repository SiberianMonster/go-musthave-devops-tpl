package storage

import (
	
	"reflect"
	"go-musthave-devops-tpl/internal/utils"
	"fmt"
	"golang.org/x/exp/slices"
	"errors"
)


func RepositoryUpdate(m utils.MetricsContainer, mp utils.UpdateMetrics ) (utils.MetricsContainer, error) {

	var v reflect.Value
	v = reflect.ValueOf(mp)
	var new_value float64
	var new_cvalue utils.Counter
	var new_gvalue utils.Gauge
	fieldName, _ := v.Field(0).Interface().(string)
	new_value, _ = v.Field(1).Interface().(float64)

	t := reflect.TypeOf(m)

	names := make([]string, t.NumField())
	for i := range names {
		names[i] = t.Field(i).Name
	}
	if slices.Contains(names, fieldName){
		if fieldName == "PollCount" {
			new_cvalue = m.PollCount + utils.Counter(new_value)
			fmt.Println(new_cvalue)
			fmt.Println(v.Field(1).Interface())
			reflect.ValueOf(&m).Elem().FieldByName(fieldName).Set(reflect.ValueOf(new_cvalue))
		} else {
			new_gvalue = utils.Gauge(new_value)
			fmt.Println(new_gvalue)
			fmt.Println(v.Field(1).Interface())
			reflect.ValueOf(&m).Elem().FieldByName(fieldName).Set(reflect.ValueOf(new_gvalue))
		}
	} else {
		return m, errors.New("missing field")
	}
	return m, nil

}





