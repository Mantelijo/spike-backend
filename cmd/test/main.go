package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/exp/rand"
)

func main() {
	updateWidgetConnectionsHttp()
}

func updateWidgetConnectionsHttp() {
	lastTs := time.Now()
	i := 0

	port := []string{"P", "R", "Q"}
	// Make sure to have that amount of ports in db a
	maxWidget := 1000
	for {
		i++
		snOwner := fmt.Sprintf("widget_sn_%d", rand.Intn(maxWidget))
		snPeer := fmt.Sprintf("widget_sn_%d", rand.Intn(maxWidget))
		nextPort := port[i%3]

		req, err := http.NewRequest("PUT", "http://localhost:8080/widgets/associations", bytes.NewBuffer(
			[]byte(
				fmt.Sprintf(`
				{
					"port_type": "%s",
					"widget_serial_num": "%s",
					"peer_widget_serial_num": "%s"
				}
`, nextPort, snOwner, snPeer),
			),
		))
		if err != nil {
			log.Fatal(err)
		}
		http.DefaultClient.Do(req)
		if i%1000 == 0 {
			fmt.Printf("Updated %d widget connections in %s\n", i, time.Since(lastTs))
			lastTs = time.Now()
		}
	}
}

func createWidgetsHttp() {
	lastTs := time.Now()
	i := 0
	for {
		i++
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
		if err != nil {
			log.Fatal(err)
		}
		http.DefaultClient.Do(req)
		if i%1000 == 0 {
			fmt.Printf("Posted %d widgets in %s\n", i, time.Since(lastTs))
			lastTs = time.Now()
		}
	}
}
