package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

func RunTestClient() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server")
	for {
		var mangaID string
		var chapter int
		fmt.Print("Enter manga ID: ")
		fmt.Scan(&mangaID)
		fmt.Print("Enter chapter number: ")
		fmt.Scan(&chapter)

		// send the progress update message
		msg := Message{
			Type: "progress_update",
			Data: map[string]any{
				"user":           "test_user",
				"manga_id":       mangaID,
				"chapter_number": chapter,
				"page":           1,
			},
		}
		bytes, _ := json.Marshal(msg)
		conn.Write(append(bytes, '\n'))
		line, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Println("Broadcast:", line)
	}
}
