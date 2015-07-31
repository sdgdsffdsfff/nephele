# Fdfs_client go
The Golang interface to the Fastdfs Ver 4.06.



## Notice

Only realized the download



###Getting started

//ip of tracker group

trackerGroup := {"10.2.25.31", "10.2.25.32"}



//tracker port

trackerPort := 22012



//fdfs client will create a connection pool to a tracker

//the tracker is selected randomly from tracker group

fdfsClient,_ := fdfs.NewFdfsClient(trackerGroup, trackerPort)



//first, the client get a connection from connetction pool, and connetct

//to the tracker to query the ip and port of a download storage server

//if the storage server doesn't exist, it will create a connection pool

//and add it to the storge server pool map, otherwise it will get directly

//from the pool map

buf,_ := fdfsClient.DownloadToBuffer("group1/M00/18/91/CgIZH1RyormAP7xTAAA1U2y_hqk858.jpg
")

