package midd

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/gin-gonic/gin"
)

var PushFile gin.HandlerFunc = func(c *gin.Context) {
	var (
		fh       *multipart.FileHeader
		err      error
		fullName string
		isSelf   bool
	)
	fh, err = c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	switch fh.Filename {
	case chart.ServerName:
		isSelf = true
		fullName = chart.RootDir + chart.ServerName
		//other file case
	default:
		fullName = filepath.Join(chart.RootDir, "upload", fh.Filename)
	}
	p, _ := filepath.Split(fullName)
	if err = os.MkdirAll(p, 0755); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	if err = c.SaveUploadedFile(fh, fullName); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "文件上传成功!"})

	//reboot?
	if isSelf {
		log.Println("self restart")
		cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s restart", os.Args[0]))
		cmd.Start()
	}
}

var Reboot gin.HandlerFunc = func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "命令发送成功!"})
	log.Println("self restart")
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s restart", os.Args[0]))
	cmd.Start()
}
