// Copyright 2018 The go-emailaddress AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package emailaddress provides a tiny library for finding, parsing and validation of email addresses.

Usage:

	import "github.com/mcnijman/go-emailaddress"

*/
package emailaddress

import (
	"fmt"
	"net"
	"net/smtp"
	"regexp"
	"strings"
)

var (
	// validEmailRegexp is a RFC 5322 regex, as per: https://stackoverflow.com/a/201378/5405453.
	// Note that this can't verify that the address is an actual working email address.
	// Use ValidateHost as a starter and/or send them one :-).
	validEmailRegexp = regexp.MustCompile("^(?i)(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])*$")
	findEmailRegexp  = regexp.MustCompile("(?i)(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])")
)

// EmailAddress is a structure that stores the address parts
type EmailAddress struct {
	LocalPart string
	Domain    string
}

func (e EmailAddress) String() string {
	return fmt.Sprintf("%s@%s", e.LocalPart, e.Domain)
}

// ValidateHost will test if the email address is actually reachable
func (e EmailAddress) ValidateHost() error {
	host, err := lookupHost(e.Domain)
	if err != nil {
		return err
	}
	return tryHost(host, e)
}

// Find uses regex to match, parse and validate any email addresses found
// in a string
func Find(haystack []byte, validateHost bool) (emails []*EmailAddress) {
	results := findEmailRegexp.FindAll(haystack, -1)
	for _, r := range results {
		if e, err := Parse(string(r)); err == nil {
			if validateHost {
				if err := e.ValidateHost(); err != nil {
					continue
				}
			}
			emails = append(emails, e)
		}
	}
	return emails
}

// Parse will parse the input and validate the email locally.
// If you want to validate this email remotely call the ValidateHost method
func Parse(email string) (*EmailAddress, error) {
	if !validEmailRegexp.MatchString(email) {
		return nil, fmt.Errorf("format is incorrect for %s", email)
	}

	i := strings.LastIndexByte(email, '@')
	e := &EmailAddress{
		LocalPart: email[:i],
		Domain:    email[i+1:],
	}
	return e, nil
}

// lookupHost first checks if any MX records are available and if not, it will check
// if A records are available as they can resolve email server hosts. An error indicates
// that non of the A or MX records are available.
func lookupHost(domain string) (string, error) {
	if mx, err := net.LookupMX(domain); err == nil {
		return mx[0].Host, nil
	}
	if ips, err := net.LookupIP(domain); err == nil {
		return ips[0].String(), nil // randomly returns IPv4 or IPv6 (when available)
	}
	return "", fmt.Errorf("failed finding MX and A records for domain %s", domain)
}

// tryHost will verify if we can start a mail transaction with the host.
func tryHost(host string, e EmailAddress) error {
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", host, 25))
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Hello(e.Domain); err == nil {
		if err = client.Mail(fmt.Sprintf("hello@%s", e.Domain)); err == nil {
			if err = client.Rcpt(e.String()); err == nil {
				client.Reset() // #nosec
				client.Quit()  // #nosec
				return nil
			}
		}
	}
	return err
}
