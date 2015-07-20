#!/bin/bash
gm_cppflags=`GraphicsMagickWand-config --cppflags`
gm_ldflags=`GraphicsMagickWand-config --ldflags`
gm_libs=`GraphicsMagickWand-config --libs`
sed -i /'#cgo LDFLAGS'/c"#cgo LDFLAGS: $gm_ldflags $gm_libs"  ../img4g/image.go
sed -i /'#cgo CPPFLAGS'/c"#cgo CPPFLAGS: $gm_cppflags" ../img4g/image.go
sed -i s/$//g ../img4g/image.go
t=$1
port=$2
nginxpath=$3
nginxport=$4
threadcount=$5

setsid go run imgsvrd.go $t $port $nginxpath $nginxport $threadcount
exit