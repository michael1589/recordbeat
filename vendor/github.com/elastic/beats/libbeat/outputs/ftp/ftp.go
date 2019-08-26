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

package ftp

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/file"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/outputs/codec"
	"github.com/elastic/beats/libbeat/publisher"
	ftpclient "github.com/jlaffaye/ftp"
)

type ftp struct {
	observer    outputs.Observer
	cfg         *config
	codec       codec.Codec
	index       string
	lock        sync.RWMutex
	fileLock    sync.Mutex
	fileRotator *file.Rotator
}

func init() {
	outputs.RegisterType("ftp", makeFtp)
}

func makeFtp(
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
	if err := config.Validate(); err != nil {
		return outputs.Fail(fmt.Errorf("ftp output config failed with: %v", err))
	}

	enc, err := codec.CreateEncoder(beat, config.Codec)
	if err != nil {
		return outputs.Fail(err)
	}

	index := beat.Beat
	c, err := newFtp(index, observer, enc, config)
	if err != nil {
		return outputs.Fail(fmt.Errorf("ftp output initialization failed with: %v", err))
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

func newFtp(index string, observer outputs.Observer, codec codec.Codec, conf *config) (*ftp, error) {
	var err error
	buf, _ := json.Marshal(conf)
	logp.Info(string(buf))

	conf.FtpFile, err = filepath.Abs(conf.FtpFile)
	if err != nil {
		logp.Err(err.Error())
	}
	mp3Rotator, err := file.NewFileRotator(
		conf.FtpFile,
		file.Interval(time.Hour*24),
		//file.MaxBackups(c.NumberOfFiles),
		//file.Permissions(os.FileMode(c.Permissions)),
		//file.WithLogger(logp.NewLogger("rotator").With(logp.Namespace("rotator"))),
	)
	if err != nil {
		return nil, err
	}

	c := &ftp{
		cfg:         conf,
		codec:       codec,
		observer:    observer,
		index:       index,
		lock:        sync.RWMutex{},
		fileLock:    sync.Mutex{},
		fileRotator: mp3Rotator,
	}
	return c, nil
}

func (c *ftp) login(url, user, password string) (*ftpclient.ServerConn, error) {
	conn, err := ftpclient.Dial(url, ftpclient.DialWithTimeout(3*time.Second))
	if err != nil {
		return nil, err
	}
	if err = conn.Login(user, password); err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *ftp) createRemoteDir(conn *ftpclient.ServerConn, file string) error {
	dir := path.Dir(file)
	if err := conn.ChangeDir(dir); err != nil {
		conn.ChangeDir("~")
	}
	baseDir, err := conn.CurrentDir()
	if err != nil {
		logp.Err(err.Error())
		return err
	}
	target := strings.Split(dir, "/")
	logp.Info(strings.Join(target, " "))
	for _, t := range target {
		baseDir = baseDir + t + "/"
		//baseDir = path.Join(baseDir, t)
		logp.Info(baseDir)
		if err := conn.ChangeDir(baseDir); err != nil {
			if err = conn.MakeDir(baseDir); err != nil {
				logp.Err(err.Error())
				return err
			}
		}
	}
	return nil
}

func (c *ftp) upload(conn *ftpclient.ServerConn, file string) error {
	if err := c.createRemoteDir(conn, file); err != nil {
		return err
	}
	f := path.Base(file)
	if err := conn.MakeDir(f); err != nil {
		logp.Err(err.Error())
		return err
	}
	return nil
}

func (c *ftp) Close() error {
	return c.fileRotator.Close()
}

func (c *ftp) Publish(batch publisher.Batch) error {

	logp.Info("Publish Begin")
	defer logp.Info("Publish End")
	st := c.observer
	events := batch.Events()
	st.NewBatch(len(events))
	limit := make(chan struct{}, c.cfg.LimitRoutine)

	conn, err := c.login(c.cfg.FtpUrl, c.cfg.FtpUser, c.cfg.FtpPasswd)
	if err != nil{
		return err
	}
	defer conn.Quit()

	wg := sync.WaitGroup{}
	for i := range events {
		limit <- struct{}{}
		wg.Add(1)
		go c.publishEvent(&events[i], &wg, limit, conn)
	}
	wg.Wait()
	batch.ACK()
	st.Acked(len(events))

	return nil
}

func (c *ftp) publishEvent(event *publisher.Event, wg *sync.WaitGroup, ch chan struct{}, conn *ftpclient.ServerConn) bool {
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
	if exists, _ := pathExists(message); !exists {
		logp.Err("%s not exists", message)
		return false
	}
	source := message
	message = strings.Replace(message, c.cfg.Prefix, c.cfg.FtpRemote, 0)
	if err = c.upload(conn, message); err != nil{
		logp.Err(err.Error())
		return false
	}

	c.fileLock.Lock()
	_, err = c.fileRotator.Write(append([]byte(message), '\n'))
	c.fileRotator.Sync()
	c.fileLock.Unlock()
	if err != nil {
		logp.Err(err.Error())
		return false
	}

	if c.cfg.DeleteSource {
		os.Remove(source)
	}
	c.observer.WriteBytes(len(serializedEvent))
	return true
}

func (c *ftp) String() string {
	return "ftp"
}
