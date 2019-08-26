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

//filebeat\vendor\github.com\elastic\beats\libbeat\publisher\includes\includes.go imports this package

package mix

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/outputs/codec"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/elastic/beats/libbeat/common/file"
)

type mix struct {
	observer outputs.Observer
	cfg      *config
	codec    codec.Codec
	index    string
	lock     sync.RWMutex
	mp3Lock	 sync.Mutex
	wavLock  sync.Mutex
	mp3Rotator	 *file.Rotator
	wavRotator	 *file.Rotator
}

type consoleEvent struct {
	Timestamp time.Time `json:"@timestamp" struct:"@timestamp"`

	// Note: stdlib json doesn't support inlining :( -> use `codec: 2`, to generate proper event
	Fields interface{} `struct:",inline"`
}

func init() {
	outputs.RegisterType("mix", makeMix)
}

func makeMix(
	_ outputs.IndexManager,
	beat beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {
	config := defaultConfig
	err := cfg.Unpack(&config)
	if err != nil {
		return outputs.Fail(err)
	}

	//var enc codec.Codec
	//if config.Codec.Namespace.IsSet() {
	enc, err := codec.CreateEncoder(beat, config.Codec)
	if err != nil {
		return outputs.Fail(err)
	}
	//} else {
	//	enc = json.New(config.Pretty, true, beat.Version)
	//}

	index := beat.Beat
	c, err := newMix(index, observer, enc, config)
	if err != nil {
		return outputs.Fail(fmt.Errorf("mix output initialization failed with: %v", err))
	}

	return outputs.Success(-1, 0, c)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func newMix(index string, observer outputs.Observer, codec codec.Codec, conf *config) (*mix, error) {
	var err error
	buf, _ := json.Marshal(conf)
	logp.Info(string(buf))
	dir := path.Dir(conf.MP3Path)
	if exist, err := pathExists(dir); !exist {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			str := fmt.Sprintf("create dir %s fail, %s", dir, err.Error())
			logp.Err(str)
			return nil, errors.New(str)
		}
	}
	dir = path.Dir(conf.WavPath)
	if exist, err := pathExists(dir); !exist {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			str := fmt.Sprintf("create dir %s fail, %s", dir, err.Error())
			logp.Err(str)
			return nil, errors.New(str)
		}
	}
	conf.MP3Path, err = filepath.Abs(conf.MP3Path)
	if err != nil{
		logp.Err(err.Error())
	}
	mp3Rotator, err := file.NewFileRotator(
		conf.MP3Path,
		file.Interval(time.Hour*24),
		//file.MaxBackups(c.NumberOfFiles),
		//file.Permissions(os.FileMode(c.Permissions)),
		//file.WithLogger(logp.NewLogger("rotator").With(logp.Namespace("rotator"))),
	)
	if err != nil {
		return nil, err
	}
	conf.WavPath, err = filepath.Abs(conf.WavPath)
	if err != nil{
		logp.Err(err.Error())
	}
	wavRotator, err := file.NewFileRotator(
		conf.WavPath,
		file.Interval(time.Hour*24),
	)
	if err != nil {
		return nil, err
	}

	c := &mix{
		cfg:      conf,
		codec:    codec,
		observer: observer,
		index:    index,
		lock:     sync.RWMutex{},
		mp3Lock:  sync.Mutex{},
		wavLock:  sync.Mutex{},
		mp3Rotator:mp3Rotator,
		wavRotator:wavRotator,
	}


	return c, nil
}

func (c *mix) Close() error {
	c.mp3Rotator.Close()
	return c.wavRotator.Close()
}

func (c *mix) Publish(batch publisher.Batch) error {
	logp.Info("Publish Begin")
	defer logp.Info("Publish End")
	st := c.observer
	events := batch.Events()
	st.NewBatch(len(events))
	limit := make(chan struct{}, c.cfg.LimitRoutine)
	//mp3, err := os.OpenFile(c.cfg.MP3Path, os.O_CREATE|os.O_APPEND, 0)
	//if err != nil{
	//	logp.Err(err.Error())
	//	return nil
	//}
	//mp3.WriteString("icsocqqqqqqqqqq====\n")
	//defer mp3.Close()
	//wav, err := os.OpenFile(c.cfg.WavPath, os.O_CREATE|os.O_APPEND, 0)
	//if err != nil{
	//	logp.Err(err.Error())
	//	return nil
	//}
	//defer wav.Close()

	wg := sync.WaitGroup{}
	for i := range events {
		limit <- struct{}{}
		wg.Add(1)
		go c.publishEvent(&events[i], &wg, limit)
	}
	wg.Wait()
	batch.ACK()
	st.Acked(len(events))

	return nil
}

func (c *mix) publishEvent(event *publisher.Event, wg *sync.WaitGroup, ch chan struct{}) bool {
	defer func() {
		<-ch
		wg.Done()
	}()
	c.lock.Lock()
	serializedEvent, err := c.codec.Encode(c.index, &event.Content)
	c.lock.Unlock()
	if err != nil {
		logp.Critical("Unable to encode event: %v", err)
		return false
	}

	content, err := event.Content.GetValue("message")
	if err != nil {
		logp.Err(err.Error())
		return false
	}
	message := content.(string)

	//deal with buffer
	//c.observer.WriteError(err) if fails
	logp.Info(message)
	if !strings.Contains(message, "res_monitor.c: ") {
		return true
	}
	m := strings.Split(message, "res_monitor.c: ")
	message = m[len(m)-1]
	params := strings.Split(message, " ")
	if len(params) < 2 {
		logp.Err("format wrong, %s", message)
		return false
	}
	In := params[0] + "-in.wav"
	out := params[0] + "-out.wav"
	//file check exists in/out
	if exist, _ := pathExists(In); !exist {
		logp.Critical("%s not exist", In)
		return false
	}
	if exist, _ := pathExists(out); !exist {
		logp.Critical("%s not exist", out)
		return false
	}

	middleDir := strings.TrimLeft(params[0], params[1])
	target := path.Join(c.cfg.MP3Dir, middleDir+".mp3")
	target, err = filepath.Abs(target)
	if err != nil{
		logp.Err(err.Error())
	}
	dir := path.Dir(target)
	if exist, err := pathExists(dir); !exist {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			logp.Err("create dir %s fail, %s", dir, err.Error())
			return false
		}
	}

	mix_wav := false
	if params[len(params)-1] != "0" {
		mix_wav = true
	}
	args := strings.Split(c.cfg.Cmd, " ")
	if len(args) == 0 {
		args = []string{"sox", "-M"}
	}

	//mp3
	args = append(args, In)
	args = append(args, out)
	args = append(args, target)
	logp.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	if err := cmd.Run(); err != nil {
		logp.Info("run cmd fail, ERROR: %s", err.Error())
		return false
	}
	logp.Info("write target: %s", target)
	c.mp3Lock.Lock()
	_, err = c.mp3Rotator.Write(append([]byte(target), '\n'))
	c.mp3Rotator.Sync()
	c.mp3Lock.Unlock()
	if err != nil{
		logp.Err(err.Error())
		return false
	}

	if mix_wav {
		target := path.Join(c.cfg.WavDir, middleDir+".wav")
		target, err = filepath.Abs(target)
		if err != nil{
			logp.Err(err.Error())
		}
		dir := path.Dir(target)
		if exist, err := pathExists(dir); !exist {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				logp.Err("create dir %s fail, %s", dir, err.Error())
				return false
			}
		}
		args := strings.Split(c.cfg.Cmd, " ")
		if len(args) == 0 {
			args = []string{"sox", "-M"}
		}

		//wav
		args = append(args, In)
		args = append(args, out)
		args = append(args, target)
		logp.Info(strings.Join(args, " "))
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			logp.Err(err.Error())
			return false
		}
		c.wavLock.Lock()
		_, err := c.wavRotator.Write(append([]byte(target), '\n'))
		c.wavRotator.Sync()
		c.wavLock.Unlock()
		if err != nil{
			logp.Err(err.Error())
			return false
		}
	}
	if c.cfg.DeleteSource {
		os.Remove(In)
		os.Remove(out)
	}
	c.observer.WriteBytes(len(serializedEvent))
	return true
}

func (c *mix) String() string {
	return "mix"
}
