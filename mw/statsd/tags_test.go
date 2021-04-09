package statsd

import (
	"strings"
	"testing"
)

func Test_constructTags(t *testing.T) {
	tagMap := map[string]string{"tag-one": "first-tag", "tag-two": "second_tag"}
	expectedTagStrings := "tag-one=first-tag,tag-two=second_tag"

	actualTagString := constructTags(tagMap)
	if !strings.Contains(actualTagString, expectedTagStrings) {
		t.Errorf("ecpected %s got %s", actualTagString, expectedTagStrings)
	}
}
