package main

import "log"

func checkErrorFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkErrorPrint(err error) {
	if err != nil {
		log.Println(err)
	}
}
