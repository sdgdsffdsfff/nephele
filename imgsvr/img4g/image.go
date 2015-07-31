package img4g

/*
#cgo CFLAGS: -std=c99
#cgo CPPFLAGS: -I/usr/local/include/GraphicsMagick
#cgo LDFLAGS: -L/usr/lib -L/usr/lib -lGraphicsMagickWand -lGraphicsMagick -ltiff -lfreetype -ljpeg -lpng16 -lXext -lSM -lICE -lX11 -llzma -lbz2 -lxml2 -lz -lm -lgomp -lpthread
#include <wand/magick_wand.h>
#include "cmagick.h"
*/
import "C"
import "unsafe"
import "errors"
import "fmt"

type Image struct {
	Format     string        // png, jpeg, bmp, gif, ...
	Blob       []byte        // raw image data
	magickWand *C.MagickWand //wand object
}

/*
CreateWand() creates a new wand for this Image by using Blob data
*/
func (this *Image) CreateWand() error {
	if this.magickWand != nil {
		this.DestoryWand()
	}
	status := C.createWand(&this.magickWand, (*C.uchar)(unsafe.Pointer(&this.Blob[0])), C.size_t(len(this.Blob)))
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		err := errors.New(fmt.Sprintf("error create magick wand: %s (ExceptionType = %d)", C.GoString(descr), etype))
		this.DestoryWand()
		return err
	}

	return nil
}

/*
DestroyWand() deallocates memory associated with this Image wand.
*/
func (this *Image) DestoryWand() {
	if this.magickWand != nil {
		C.DestroyMagickWand(this.magickWand)
		this.magickWand = (*C.MagickWand)(nil)
	}
}

/*
Resize() resizes the size of this image to the given dimensions.

width: width of the resized image
height: height of the resized image
*/

func (this *Image) Resize(width int64, height int64) error {
	if this.magickWand == nil {
		return errors.New("error resizing image:magickwand is nil")
	}

	status := C.MagickResizeImage(this.magickWand, C.ulong(width), C.ulong(height), C.CubicFilter, 0.5)
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error resizing image: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Composite() composite one image onto another at the specified offset.

compositeImg: The composite image
x: The column offset of the composited image.
y: The row offset of the composited image.
*/
func (this *Image) Composite(compositeImg *Image, x int64, y int64) error {
	if this.magickWand == nil {
		return errors.New("error composite image:magickwand is nil")
	}

	if compositeImg.magickWand == nil {
		return errors.New("error composite image:composite image wand is nil")
	}

	compositeImg.Dissolve(100)

	status := C.MagickCompositeImage(this.magickWand, compositeImg.magickWand, C.OverCompositeOp, C.long(x), C.long(y))
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error composite image: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Crop() extracts a region of this image.

width: the region width
height: the region height
x: the region x offset
y: the region y offset
*/
func (this *Image) Crop(width int64, height int64, x int64, y int64) error {
	if this.magickWand == nil {
		return errors.New("error crop image:magickwand is nil")
	}

	status := C.MagickCropImage(this.magickWand, C.ulong(width), C.ulong(height), C.long(x), C.long(y))
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error crop image: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Rotate() rotates an image the specified number of degrees.

degrees: degrees of the rotated image
*/
func (this *Image) Rotate(degrees float64) error {
	if this.magickWand == nil {
		return errors.New("error rotate image:magickwand is nil")
	}

	status := C.rotateImage(this.magickWand, C.double(degrees))
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error rotate image: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Scale() scales the size of this image to the given dimensions.

columns: The number of columns in the scaled image.
rows: The number of rows in the scaled image
*/
func (this *Image) Scale(columns int64, rows int64) error {
	if this.magickWand == nil {
		return errors.New("error scale image:magickwand is nil")
	}

	status := C.MagickScaleImage(this.magickWand, C.ulong(columns), C.ulong(rows))
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error scale image: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Dissovle() sets transparency of this image to the specified value dissolve

dissolve: 0~100,0 means totally transparent while 100 means opa
*/
func (this *Image) Dissolve(dissolve int) error {
	if this.magickWand == nil {
		return errors.New("error dissolve image:magickwand is nil")
	}

	status := C.dissolveImage(this.magickWand, C.uint(dissolve))
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error scale image: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Sets the image quality factor, which determines compression options when saving the file

quality: The image quality
*/
func (this *Image) SetCompressionQuality(quality int) error {
	if this.magickWand == nil {
		return errors.New("error set image compression quality:magickwand is nil")
	}

	status := C.MagickSetCompressionQuality(this.magickWand, C.ulong(quality))
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error set image compression quality: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Sets the file or blob format (e.g. "BMP") to be used when a file or blob is read. Usually this is
not necessary because GraphicsMagick is able to auto-detect the format based on the file header
(or the file extension), but some formats do not use a unique header or the selection may be ambigious.
 Use MagickSetImageFormat() to set the format to be used when a file or blob is to be written.

format: The file or blob format
*/
func (this *Image) SetFormat(format string) error {
	if this.magickWand == nil {
		return errors.New("error set image format:magickwand is nil")
	}

	var cs *C.char = C.CString(format)
	defer C.free(unsafe.Pointer(cs))
	status := C.MagickSetImageFormat(this.magickWand, cs)
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error set image format: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

/*
Returns the image height
*/
func (this *Image) GetHeight() (int64, error) {
	if this.magickWand == nil {
		return 0, errors.New("error get image height:magickwand is nil")
	}

	height, err := C.MagickGetImageHeight(this.magickWand)

	return int64(height), err
}

/*
Returns the image width
*/
func (this *Image) GetWidth() (int64, error) {
	if this.magickWand == nil {
		return 0, errors.New("error get image height:magickwand is nil")
	}

	width, err := C.MagickGetImageWidth(this.magickWand)

	return int64(width), err
}

/*
Strip() removes all profiles and text attributes from this image.
*/
func (this *Image) Strip() error {
	if this.magickWand == nil {
		return errors.New("error strip image:magickwand is nil")
	}

	status := C.MagickStripImage(this.magickWand)
	if status == 0 {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		return errors.New(fmt.Sprintf("error strip image: %s (ExceptionType = %d)", C.GoString(descr), etype))
	}

	return nil
}

func (this *Image) Size() (int64, int64, error) {
	var (
		w   int64
		h   int64
		err error
	)
	if w, err = this.GetWidth(); err != nil {
		return 0, 0, err
	}
	if h, err = this.GetHeight(); err != nil {
		return 0, 0, err
	}
	return w, h, nil
}

/*
WriteImageBlob() writes this image wand to Blob
*/
func (this *Image) WriteImageBlob() error {
	if this.magickWand == nil {
		return errors.New("error write image to blob:magickwand is nil")
	}
	var sizep int = 0

	blob := C.MagickWriteImageBlob(this.magickWand, (*C.size_t)(unsafe.Pointer(&sizep)))
	if blob != nil {
		defer C.free(unsafe.Pointer(blob))
	} else {
		var etype int
		descr := C.MagickGetException(this.magickWand, (*C.ExceptionType)(unsafe.Pointer(&etype)))
		defer C.MagickRelinquishMemory(unsafe.Pointer(descr))
		err := errors.New(fmt.Sprintf("error write image to blob: %s (ExceptionType=%d)", C.GoString(descr), etype))
		return err
	}

	this.Blob = C.GoBytes(unsafe.Pointer(blob), C.int(sizep))

	return nil
}
