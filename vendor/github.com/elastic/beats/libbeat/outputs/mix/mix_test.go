package mix

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"sync"
	"testing"
)

func TestMix(t *testing.T) {
	m, err := newMix("name", nil, nil, defaultConfig)
	assert.NotNil(t, m.lock)
	assert.NoError(t, err)
}

func TestBufio(t *testing.T) {
	mp3File, err := os.OpenFile("D://abcd.txt", os.O_CREATE|os.O_APPEND, 0)
	assert.NoError(t, err)
	//mp3 := bufio.NewWriter(mp3File)
	//mp3File.Close()
	//mp3File.WriteString("aaa\n")
	//mp3File.WriteString("bbb\n")
	//mp3.Flush()
	ch := make(chan struct{}, 3)
	for i := 0; i<3; i++{
		go procfunc(mp3File, ch, i)
	}
	for i := 0; i<3; i++ {
		<-ch
	}
	mp3File.Close()
}

func procfunc(f *os.File,ch chan struct{}, i int){
	defer func() {
		ch<- struct{}{}
	}()
	f.WriteString(fmt.Sprintf("cccc%d\n", i))
}

func TestFor(t *testing.T) {
	wg := sync.WaitGroup{}
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := range a {
		wg.Add(1)
		go func(ii *int) {
			defer wg.Done()
			fmt.Println(*ii)
		}(&a[i])
	}
	wg.Wait()
}

func TestCmd(t *testing.T) {
	args := []string{"echo", "-m",
		"/var/record/call/20190815/9/72/20190815134904-6567643136615518208-6567643136615518209-07942024805-1000-in.wav",
		"/var/record/call/20190815/9/72/20190815134904-6567643136615518208-6567643136615518209-07942024805-1000-out.wav",
		"/usr/local/filebeat/mix/mp3Dir/20190815/9/72/20190815134904-6567643136615518208-6567643136615518209-07942024805-1000.mp3",
	}
	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Run()
	assert.NoError(t, err)

}
