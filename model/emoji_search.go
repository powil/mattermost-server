// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"

	jsoniter "github.com/json-iterator/go"
)

type EmojiSearch struct {
	Term       string `json:"term"`
	PrefixOnly bool   `json:"prefix_only"`
}

func (es *EmojiSearch) ToJson() string {
	b, _ := jsoniter.Marshal(es)
	return string(b)
}

func EmojiSearchFromJson(data io.Reader) *EmojiSearch {
	var es *EmojiSearch
	json.NewDecoder(data).Decode(&es)
	return es
}
