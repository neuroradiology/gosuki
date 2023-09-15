// vim: set filetype=go noexpandtab ts=2 sts=2 sw=2:
// Package database is
package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


// PreSanitize the list of tags before saving them to the DB
// It should replace the delim with a double dash
func TestTags_PreSanitize(t *testing.T) {
	type fields struct {
		delim string
		tags  []string
	}
	tests := []struct {
		name   string
		fields fields
		want   Tags
	}{
		{"empty", fields{"", []string{}}, Tags{"", []string{}}},
		{"good_input", fields{",", []string{"tag1", "tag2"}}, Tags{",", []string{"tag1", "tag2"}}},
		{"bad_input1", fields{",", []string{"tag1,", "tag2"}}, Tags{",", []string{"tag1--", "tag2"}}},
		{"bad_input2", fields{",", []string{"tag1", ",tag2"}}, Tags{",", []string{"tag1", "--tag2"}}},
		{"bad_input3", fields{",", []string{"tag1,", ",tag2"}}, Tags{",", []string{"tag1--", "--tag2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Tags{
				delim: tt.fields.delim,
				tags:  tt.fields.tags,
			}
			// if got := tr.PreSanitize(); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Tags.PreSanitize() = %v, want %v", got, tt.want)
			// }
			assert.Equal(t, tt.want, *(tr.PreSanitize()))
		})
	}
}

func TestTags_StringWrap(t *testing.T) {
	type fields struct {
		delim string
		tags  []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Tags{
				delim: tt.fields.delim,
				tags:  tt.fields.tags,
			}
			if got := tr.StringWrap(); got != tt.want {
				t.Errorf("Tags.StringWrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagsFromString(t *testing.T) {
	delim := ","
	type args struct {
		s     string
	}
	tests := []struct {
		name string
		tagstr string
		want []string
	}{
		{"case1", "tag1,tag2", []string{"tag1", "tag2"}},
		{"case2", ",tag1,tag2,,", []string{"tag1", "tag2"}},
		{"case2", ",,tag1,,,tag2,,tag3,,", []string{"tag1", "tag2", "tag3"}},
		{"empty", "", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.Equal(t, TagsFromString(tt.tagstr, delim).tags, tt.want)
		})
	}
}

func Test_delimWrap(t *testing.T) {
	type args struct {
		token string
		delim string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", ","}, ","},
		{"non wrapped", args{"tag1,tag2", ","}, ",tag1,tag2,"},
		{"wrapped", args{",tag1,tag2,", ","}, ",tag1,tag2,"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, delimWrap(tt.args.token, tt.args.delim))
		})
	}
}
