package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	for i := range 50_000 {
		name := fmt.Sprintf("widget_name_%d", i)
		sn := fmt.Sprintf("widget_sn_%d", i)
		req, err := http.NewRequest("POST", "http://localhost:8080/widgets", bytes.NewBuffer(
			[]byte(
				fmt.Sprintf(`
			{
				"name": "%s",
				"serial_number": "%s",
				"ports": ["P", "R", "Q"]
			}`, name, sn),
			),
		))
		log.Fatal(err)
		http.DefaultClient.Do(req)
	}
}
