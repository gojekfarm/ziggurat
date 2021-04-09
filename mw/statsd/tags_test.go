package statsd

import "testing"

func Test_constructTags(t *testing.T) {
	tagMap := map[string]string{"tag-one": "first-tag", "tag-two": "second_tag"}
	expectedTagString := "tag-one=first-tag,tag-two=second_tag"
	actualTagString := constructTags(tagMap)
	if expectedTagString != actualTagString {
		t.Errorf("expected %s, got %s", expectedTagString, actualTagString)
	}
}
