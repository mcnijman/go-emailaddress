// Copyright 2018 The go-emailaddress AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package emailaddress

import (
	"reflect"
	"testing"
)

func TestEmailAddress_String(t *testing.T) {
	type fields struct {
		LocalPart string
		Domain    string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"1", fields{"foo", "bar.com"}, "foo@bar.com"},
		{"2", fields{"foo", ""}, ""},
		{"3", fields{"", "bar.com"}, ""},
		{"4", fields{"", ""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := EmailAddress{
				LocalPart: tt.fields.LocalPart,
				Domain:    tt.fields.Domain,
			}
			if got := e.String(); got != tt.want {
				t.Errorf("EmailAddress.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailAddress_ValidateHost(t *testing.T) {
	type fields struct {
		LocalPart string
		Domain    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"1", fields{"fake", "example.com"}, true},
		{"2", fields{"fake", "foo.foobar"}, true},
		{"3", fields{"infos", "google.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := EmailAddress{
				LocalPart: tt.fields.LocalPart,
				Domain:    tt.fields.Domain,
			}
			if err := e.ValidateHost(); (err != nil) != tt.wantErr {
				t.Errorf("EmailAddress.ValidateHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailAddress_ValidateIcanSuffix(t *testing.T) {
	type fields struct {
		LocalPart string
		Domain    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"1", fields{"fake", "example.com"}, false},
		{"2", fields{"fake", "foo.foobar"}, true},
		{"3", fields{"info", "google.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := EmailAddress{
				LocalPart: tt.fields.LocalPart,
				Domain:    tt.fields.Domain,
			}
			if err := e.ValidateIcanSuffix(); (err != nil) != tt.wantErr {
				t.Errorf("EmailAddress.ValidateIcanSuffix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFind(t *testing.T) {
	type args struct {
		haystack       []byte
		validateRemote bool
	}
	tests := []struct {
		name       string
		args       args
		wantEmails []*EmailAddress
	}{
		{"1", args{[]byte(`test@example.com`), false}, []*EmailAddress{{"test", "example.com"}}},
		{"2", args{[]byte(`Sample text test@example.com.`), false}, []*EmailAddress{{"test", "example.com"}}},
		{"3", args{[]byte(`Sample text TestEmail@Example.com.`), false}, []*EmailAddress{{"TestEmail", "Example.com"}}},
		{"4", args{[]byte(`Send me an email at this@domain.com or info@domain.com or not.`), false}, []*EmailAddress{{"this", "domain.com"}, {"info", "domain.com"}}},
		{"5", args{[]byte(`Send me an email at fake@example.com.`), true}, nil},
		{"6", args{[]byte(`<ul><li>Joe Smith has moved on to<a href="http://www.Google.com/">Google</a>, 1600 Amphitheatre Parkway,Mountain View, CA 94043</li><li>info9@google.com</li></ul>`), true}, []*EmailAddress{{"info9", "google.com"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotEmails := Find(tt.args.haystack, tt.args.validateRemote); !reflect.DeepEqual(gotEmails, tt.wantEmails) {
				t.Errorf("Find() = %v, want %v", gotEmails, tt.wantEmails)
			}
		})
	}
}

func TestFindWithIcannSuffix(t *testing.T) {
	type args struct {
		haystack     []byte
		validateHost bool
	}
	tests := []struct {
		name       string
		args       args
		wantEmails []*EmailAddress
	}{
		{"1", args{[]byte(`Sample text test@example.com.`), false}, []*EmailAddress{{"test", "example.com"}}},
		{"2", args{[]byte(`Sample text test@example.foobar.`), false}, nil},
		{"3", args{[]byte(`Send me an email at fake@example.foobar.`), true}, nil},
		{"4", args{[]byte(`<ul><li>Joe Smith has moved on to<a href="http://www.Google.com/">Google</a>, 1600 Amphitheatre Parkway,Mountain View, CA 94043</li><li>info10@google.com</li></ul>`), true}, []*EmailAddress{{"info10", "google.com"}}},
		{"5", args{[]byte(`Sample text test@25c95f9e-b0d4-4d67-a159-56f360b48273.museum.`), true}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotEmails := FindWithIcannSuffix(tt.args.haystack, tt.args.validateHost); !reflect.DeepEqual(gotEmails, tt.wantEmails) {
				t.Errorf("FindWithIcannSuffix() = %v, want %v", gotEmails, tt.wantEmails)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    *EmailAddress
		wantErr bool
	}{
		{"valid_1", args{"email@domain.com"}, &EmailAddress{"email", "domain.com"}, false},
		{"valid_2", args{"firstname+last.name@domain.com"}, &EmailAddress{"firstname+last.name", "domain.com"}, false},
		{"valid_3", args{"\"email\"@domain.com"}, &EmailAddress{"\"email\"", "domain.com"}, false},
		{"valid_4", args{"1234567890@domain.com"}, &EmailAddress{"1234567890", "domain.com"}, false},
		{"valid_5", args{"_______@domain.com"}, &EmailAddress{"_______", "domain.com"}, false},
		{"valid_6", args{"email@domain.museum"}, &EmailAddress{"email", "domain.museum"}, false},
		{"valid_7", args{"email@sub.domain.co.uk"}, &EmailAddress{"email", "sub.domain.co.uk"}, false},
		{"valid_8", args{"firstname-lastname@domain.com"}, &EmailAddress{"firstname-lastname", "domain.com"}, false},
		{"valid_9", args{"email@123.123.123.123"}, &EmailAddress{"email", "123.123.123.123"}, false},
		{"valid_10", args{"email@[123.123.123.123]"}, &EmailAddress{"email", "[123.123.123.123]"}, false},
		{"valid_11", args{"FirstNameLastName@domain.com"}, &EmailAddress{"FirstNameLastName", "domain.com"}, false},
		{"valid_12", args{"FirstNameLastName@doMain.com"}, &EmailAddress{"FirstNameLastName", "doMain.com"}, false},
		{"invalid_1", args{"plainaddress"}, nil, true},
		{"invalid_2", args{"#@%^%#$@#$@#.com"}, nil, true},
		{"invalid_3", args{"@domain.com"}, nil, true},
		{"invalid_4", args{"Joe Smith <email@domain.com>"}, nil, true},
		{"invalid_5", args{"email.domain.com"}, nil, true},
		{"invalid_6", args{"email@domain@domain.com"}, nil, true},
		{"invalid_7", args{".email@domain.com"}, nil, true},
		{"invalid_8", args{"email.@domain.com"}, nil, true},
		{"invalid_9", args{"email..email@domain.com"}, nil, true},
		{"invalid_10", args{"あいうえお@domain.com"}, nil, true},
		{"invalid_11", args{"email@domain.com (Joe Smith)"}, nil, true},
		{"invalid_12", args{"email@domain"}, nil, true},
		{"invalid_13", args{"email@-domain.com"}, nil, true},
		{"invalid_14", args{"email@domain..com"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_LookupHost(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{"google.com"}, false},
		{"2", args{"example.com"}, false},
		{"3", args{"fake.foobar"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LookupHost(tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" && !tt.wantErr {
				t.Errorf("LookupHost() = %v, want non empty", got)
			}
		})
	}
}

func Test_TryHost(t *testing.T) {
	type args struct {
		host string
		e    EmailAddress
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{"aspmx.l.google.com.", EmailAddress{"info1", "google.com"}}, false},
		{"2", args{"173.194.68.27", EmailAddress{"info2", "google.com"}}, false},
		{"3", args{"non valid host", EmailAddress{"fake", "example.com"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TryHost(tt.args.host, tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("TryHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
