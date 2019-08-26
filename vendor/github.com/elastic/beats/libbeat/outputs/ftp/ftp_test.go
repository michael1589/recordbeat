package ftp
import (
	"github.com/stretchr/testify/assert"
	"testing"
)
func TestUpload(t *testing.T) {
	f := ftp{}
	conn, err := f.login("172.25.1.138:21", "ftprec_test", "test_jfwe852Fdew")
	assert.NoError(t, err)
	err = f.upload(conn, "E:/book/quants basic/3/src/boll_factor.py")
	assert.NoError(t, err)
}