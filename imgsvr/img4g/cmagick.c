#include <wand/magick_wand.h>
#include  <magick/attribute.h>
#include <magick/blob.h>     
#include <magick/error.h>    
#include <magick/image.h>
#include <magick/list.h>    
#include <magick/paint.h>    
#include <magick/quantize.h> 
#include <magick/resize.h>
#include <magick/resource.h> 
#include <magick/api.h>    

#include "cmagick.h"

typedef struct _MagickWand
{
  char
    id[MaxTextExtent];

  ExceptionInfo
    exception;

  ImageInfo
    *image_info;

  QuantizeInfo
    *quantize_info;

  Image
    *image,             /* Current working image */
    *images;            /* Whole image list */

  unsigned int
    iterator;

  unsigned long
    signature;
}NewWand;



unsigned int dissolveImage(MagickWand *wand, const unsigned int dissolve)
{
    int x,y;
    register PixelPacket *q;
    NewWand *newWand;
    Image *image;
    
    newWand = (NewWand *)wand;
    image = newWand->image;
    if (!image->matte)
        SetImageOpacity(image,OpaqueOpacity);
    for(y=0; y< (long)image->rows; y++)
    {
        q=GetImagePixels(image,0,y,image->columns,1);
        if (q == (PixelPacket *) NULL)
        {
            CopyException(&newWand->exception, &image->exception);
            return(False);
        }
        for (x=0; x < (long) image->columns; x++)
        {
            if(q->opacity != MaxRGB)
            {
                q->opacity=(Quantum)(MaxRGB - ((MaxRGB-q->opacity)/100.0*dissolve));
            }
            q++;
        }
        if (!SyncImagePixels(image))
        {
            CopyException(&newWand->exception, &image->exception);
            return(False);
        }
    }
    return(True);
}

unsigned int rotateImage(MagickWand *wand, double degrees)
{
    PixelWand *background;
    unsigned int status;

    background = NewPixelWand();
    status = PixelSetColor(background, "#000000");
    
    if (status == True)
        status =  MagickRotateImage(wand, background, degrees);

    DestroyPixelWand(background);
    return status;
}

unsigned int createWand(MagickWand **wand,const unsigned char *blob,const size_t length)
{
    *wand = NewMagickWand();
    return MagickReadImageBlob(*wand, blob, length);
}
