package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	suffix := "." + domain

	var u struct {
		Email string
	}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := json.Unmarshal(line, &u); err != nil {
			return nil, fmt.Errorf("parse error: %w", err)
		}
		email := strings.ToLower(u.Email)
		if strings.HasSuffix(email, suffix) {
			if at := strings.IndexByte(email, '@'); at >= 0 {
				result[email[at+1:]]++
			}
		}
	}

	return result, scanner.Err()
}
