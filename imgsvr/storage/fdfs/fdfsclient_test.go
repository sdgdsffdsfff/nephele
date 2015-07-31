package fdfs

import (
	"testing"
//	"time"
) 

func TestDownloadToBuffer(t *testing.T) {
	println("enter test")
	trackerIp := []string{"10.2.25.31"}
	trackerPort := "22122"
	fdfsClient, err := NewFdfsClient(trackerIp, trackerPort)
	if err != nil {
		t.Error(err)
	}
	_, err = fdfsClient.DownloadToBuffer("group1/M00/18/91/CgIZH1RyormAP7xTAAA1U2y_hqk858.jpg")
	if err != nil {
		t.Error(err)
	} 
	for i:=0; i<200; i++ {
		go func() {
			for {
				_, err := fdfsClient.DownloadToBuffer("group1/M00/18/91/CgIZH1RyormAP7xTAAA1U2y_hqk858.jpg")
				if err != nil {
					println("download error", err.Error())
				} else {
					println("success")
				}

			}
		}()
	}
	c := make (chan int)
	<- c
	
}
