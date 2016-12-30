package irgen

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestIrgen(t *testing.T) {
	example := exec.Command("lli-3.8")

	example.Stdin = strings.NewReader(`
        define i32 @main() {
            entry:
                ret i32 123
        }
    `)

	var out bytes.Buffer
	example.Stdout = &out
	example.Stderr = &out

	if err := example.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			re := regexp.MustCompile("[0-9]+")
			code, _ := strconv.Atoi(re.FindAllString(exitErr.Error(), -1)[0])
			log.Println(code)
		}
		log.Fatal(err)
	}

	fmt.Println(out.String())
}
