package message

import (
	"fmt"

	"github.com/faint/server"
)

func testMain() {
	s := server.GetServer(20202)
	fmt.Println(s)
}
