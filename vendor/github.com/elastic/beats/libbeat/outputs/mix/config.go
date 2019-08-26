// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package mix

import (
	"github.com/elastic/beats/libbeat/outputs/codec"
)

type config struct {
	MP3Path          string       `config:"mp3path"`
	MP3Dir          string       `config:"mp3dir"`
	WavPath      string       `config:"wavpath"`
	WavDir      string       `config:"wavdir"`
	DeleteSource bool	`config:"deletesource"`
	Codec         codec.Config `config:"codec"`
	LimitRoutine	uint	`config:"limitroutine"`
	Cmd 		string	`config:"cmd"`
}

var (
	defaultConfig = &config{
		MP3Path:"./mixMp3/mp3.txt",
		MP3Dir:"/sas/rec/call_mp3",
		WavPath:"./mixWav/wav.txt",
		WavDir:"/sas/rec/call_wav",
		DeleteSource:false,
		LimitRoutine:uint(5),
	}
)

func (c *config) Validate() error {
	return nil
}
