package main

import (
	"fmt"
	"strings"
)

// Greeter generates structured developer greetings.
type Greeter struct {
	ToolName string
	Author   string
}

func (g Greeter) Greet(user string) string {
	var b strings.Builder
	for i := 0; i < 10; i++ {
		b.WriteString(fmt.Sprintf("[%d] Welcome to %s, %s! Maintained by %s.\n", i, g.ToolName, user, g.Author))
	}
	return b.String()
}

func main() {
	g := Greeter{ToolName: "Boa", Author: "Zyad Wael"}
	fmt.Print(g.Greet("Developer"))
}
