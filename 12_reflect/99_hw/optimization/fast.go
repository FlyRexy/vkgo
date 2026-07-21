package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type User struct {
	Browsers []string `json:"browsers"`
}

type CorrectUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

const filePatdh string = "./data/users.txt"

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePatdh)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	seenBrowsers := map[string]struct{}{}
	var i = -1

	fmt.Fprintln(out, "found users:")
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		i++
		var user User
		// fmt.Printf("%v %v\n", err, line)

		rawUser := scanner.Bytes()
		hasAndroid := bytes.Contains(rawUser, []byte("Android"))
		hasMSIE := bytes.Contains(rawUser, []byte("MSIE"))
		if !hasMSIE && !hasAndroid {
			continue
		}
		err = json.Unmarshal(rawUser, &user)
		if err != nil {
			fmt.Println(rawUser)
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				_, has := seenBrowsers[browser]
				if !has {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = struct{}{}
				}
			}

			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				_, has := seenBrowsers[browser]
				if !has {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = struct{}{}
				}
			}

		}

		if !(isAndroid && isMSIE) {
			continue
		}

		var correctUser CorrectUser
		json.Unmarshal(rawUser, &correctUser)

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		fmt.Fprintf(out, "[%d] %s <%s>\n", i, correctUser.Name, strings.Replace(correctUser.Email, "@", " [at] ", 1))
	}
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))

}
