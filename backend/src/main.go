package main

import (
	"hr-dynamic-form-engine/src/cfg"
	"hr-dynamic-form-engine/src/dao"
	"hr-dynamic-form-engine/src/log"
	"hr-dynamic-form-engine/src/rdb"
	"hr-dynamic-form-engine/src/router"
)

func main() {
	log.Init()
	cfg.Init()
	dao.Init()
	rdb.Init()
	router.Init()
}
