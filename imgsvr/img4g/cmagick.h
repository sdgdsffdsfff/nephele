#include <stdlib.h>
#include <wand/magick_wand.h>  


#define True 1
#define False 0

extern unsigned int dissolveImage(MagickWand *, const unsigned int);
extern unsigned int rotateImage(MagickWand *, double);
extern unsigned int createWand(MagickWand **,const unsigned char *,const size_t);


