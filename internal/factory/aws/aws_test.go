package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"reflect"
	"testing"
)

func TestMapToTagging(t *testing.T) {
	type args struct {
		m map[string]string
	}
	tests := []struct {
		name string
		args args
		want *s3.Tagging
	}{
		{
			name: "valid map",
			args: args{
				m: map[string]string{
					"Name":   "test-bucket",
					"author": "s3-operator",
				},
			},
			want: &s3.Tagging{TagSet: []*s3.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-bucket"),
				},
				{
					Key:   aws.String("author"),
					Value: aws.String("s3-operator"),
				},
			}},
		},
		{
			name: "empty map",
			args: args{
				m: map[string]string{},
			},
			want: &s3.Tagging{TagSet: []*s3.Tag{}},
		},
		{
			name: "nil map",
			args: args{
				m: nil,
			},
			want: &s3.Tagging{TagSet: []*s3.Tag{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapToTagging(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToTagging() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagListToTagMap(t *testing.T) {
	type args struct {
		tSet []*s3.Tag
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "valid tag set to map",
			args: args{
				tSet: []*s3.Tag{
					{
						Key:   aws.String("testKey"),
						Value: aws.String("testValue"),
					},
				},
			},
			want: map[string]string{"testKey": "testValue"},
		},
		{
			name: "empty tag set to map",
			args: args{
				tSet: []*s3.Tag{},
			},
			want: map[string]string{},
		},
		{
			name: "nil tag set to map",
			args: args{
				tSet: nil,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TagListToTagMap(tt.args.tSet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TagListToTagMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
