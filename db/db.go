package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func getURL() string {
	port, err := strconv.Atoi(os.Getenv("DATABASE_PORT"))
	if err != nil {
		log.Println("error on load db port form env:", err.Error())
		port = 27018
	}
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/#%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASS"),
		os.Getenv("DATABASE_HOST"),
		port,
		os.Getenv("DATABASE_NAME"))
}
