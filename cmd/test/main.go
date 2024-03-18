package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/artilugio0/vorl"
)

func main() {
	repl, err := vorl.NewREPL(interpreter{}, "vor >", "")
	if err != nil {
		panic(err)
	}

	args := os.Args
	if len(args) > 1 {
		if err := repl.RunNonInteractive(args[1]); err != nil {
			panic(err)
		}
		return
	}

	if err := repl.Run(); err != nil {
		panic(err)
	}
}

type interpreter struct{}

func (i interpreter) Exec(input string) (interface{}, error) {
	if input == "test" {
		return vorl.CommandResultSimple(`POST /v1/auth/initialize HTTP/1.1
Content-Type: application/json
Connection: keep-alive
Accept: application/json, text/plain, */*
User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/116.0
Content-Length: 124
Host: pixels-server.pixels.xyz
Sec-Fetch-Dest: empty
Referer: https://play.pixels.xyz/
Origin: https://play.pixels.xyz
Sec-Fetch-Site: same-site
Sec-Fetch-Mode: cors
Accept-Language: en-US,en;q=0.5

{"authToken":"TCKohTpjOkcG59Klt4JMA1VVT0Rjqpw_y-IBXzbunge9","mapId":"","tenant":"pixels","walletProvider":"ronin","ver":6.6}`), nil
	}
	if input == "lista" {
		return vorl.CommandResultList{
			List: []string{"a", "b", "bc", "c", "cd"},
		}, nil
	}

	if input == "error" {
		return nil, fmt.Errorf("this is an error!!!")
	}

	return vorl.CommandResultTable{
		Table: [][]string{
			{"n", "col1", "col2"},
			{"101", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "b"},
			{"12", "c", "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"},
			{"1200", "z", "xxxx"},
		},
		OnSelect: func([]string) interface{} {
			return vorl.CommandResultSimple(strings.Repeat("0123456789", 100) + "#")
		},
	}, nil
}

func (i interpreter) Suggest(string) []string {
	return nil
}
