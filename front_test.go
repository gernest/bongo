package bongo

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

var jsonPost = `+++
{
	"title":"bongo"
}
+++

# Body
Over my dead body`

var yamlPost = `---
title: chapter two
section: blog
---
A brave new world
`

func TestJSONMatter(t *testing.T) {
	m := NewJSON("+++")

	f, b, err := m.Parse(strings.NewReader(jsonPost))
	if err != nil {
		t.Error(err)
	}
	if b == nil {
		t.Error("expected body got nil instead")
	}
	if f == nil {
		t.Fatal("expecetd front")
	}
	if _, ok := f["title"]; !ok {
		t.Error("expeced title")
	}
}

func TestYAMLMatter(t *testing.T) {
	m := NewYAML("---")
	f, b, err := m.Parse(strings.NewReader(yamlPost))
	if err != nil {
		t.Error(err)
	}
	if b == nil {
		t.Error("expected body got nil instead")
	}
	if f == nil {
		t.Fatal("expecetd front")
	}
	if _, ok := f["title"]; !ok {
		t.Error("expeced title")
	}

	//	body, _ := ioutil.ReadAll(b)
	//	t.Error(string(body))
}

func TestLargeFile(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/data/post.md")
	if err != nil {
		t.Error(err)
	}
	m := NewYAML("---")
	f, b, err := m.Parse(bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}
	if b == nil {
		t.Error("expected body got nil instead")
	}
	if f == nil {
		t.Fatal("expecetd front")
	}
	if _, ok := f["title"]; !ok {
		t.Error("expeced title")
	}
	//	body, _ := ioutil.ReadAll(b)
	//	t.Error(string(body))
}
