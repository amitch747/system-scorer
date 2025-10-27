package collector

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"testing"
)

func Test_entires(t *testing.T) {
	entries, _ := os.ReadDir("/run/user")
	for _, e := range entries {
		uid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		user, err := user.LookupId(strconv.Itoa(uid))
		if err != nil {
			continue
		}
		fmt.Println(user.Username)
		user.Username = "ball"
	}
}
