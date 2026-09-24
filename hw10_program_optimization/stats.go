package hw10programoptimization

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// User represents one single user from JSON.
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

// UserEmail mapping into email.
type UserEmail struct {
	Email string
}

// DomainStat represents statistic like "domain-count".
type DomainStat map[string]int

// GetDomainStat read from source and get stat by domain.
func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	re := regexp.MustCompile("\\." + domain)
	for scanner.Scan() {
		var email UserEmail

		if err := email.UnmarshalJSON(scanner.Bytes()); err != nil {
			return nil, err
		}

		matched := re.MatchString(email.Email)

		if matched {
			num := result[strings.ToLower(strings.SplitN(email.Email, "@", 2)[1])]
			num++
			result[strings.ToLower(strings.SplitN(email.Email, "@", 2)[1])] = num
		}
	}

	return result, scanner.Err()
}
