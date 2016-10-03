package main

import (
	"fmt"
	"image"
	"image/color"
)

// draw face rectangle on image with color c
func DrawFaceRectangle(i *image.RGBA, f FaceRectangle, c color.RGBA) {
	for x := f.Left; x < f.Left+f.Width; x++ {
		i.Set(x, f.Top, c)
		i.Set(x, f.Top+1, c)
		i.Set(x, f.Top+2, c)
	}
	for x := f.Left; x < f.Left+f.Width; x++ {
		i.Set(x, f.Top+f.Height, c)
		i.Set(x, f.Top+f.Height-1, c)
		i.Set(x, f.Top+f.Height-2, c)
	}
	for y := f.Top; y < f.Top+f.Height; y++ {
		i.Set(f.Left, y, c)
		i.Set(f.Left+1, y, c)
		i.Set(f.Left+2, y, c)
	}
	for y := f.Top; y < f.Top+f.Height; y++ {
		i.Set(f.Left+f.Width, y, c)
		i.Set(f.Left+f.Width-1, y, c)
		i.Set(f.Left+f.Width-2, y, c)
	}
}

// draw numbers [1...9] on images to indicates emotions at left bottom corner of face rectangle
func DrawNumberOnImage(im *image.RGBA, number int, f FaceRectangle, c color.RGBA) {
	switch number {
	case 1:
		/*
			18	XXXXXX
			17	XXXXXX
			16	XXXXXX
			15	   XXX
			14	   XXX
			13	   XXX
			12	   XXX
			11	   XXX
			10	   XXX
			9	XXXXXXXXX
			8	XXXXXXXXX
			7	XXXXXXXXX
		*/

		for i := 6; i <= 14; i++ {
			im.Set(f.Left+i, f.Top+f.Height-7, c)
			im.Set(f.Left+i, f.Top+f.Height-8, c)
			im.Set(f.Left+i, f.Top+f.Height-9, c)
		}

		for j := 10; j <= 18; j++ {
			im.Set(f.Left+9, f.Top+f.Height-j, c)
			im.Set(f.Left+10, f.Top+f.Height-j, c)
			im.Set(f.Left+11, f.Top+f.Height-j, c)
		}

		for i := 6; i <= 8; i++ {
			for j := 16; j <= 18; j++ {
				im.Set(f.Left+i, f.Top+f.Height-j, c)
			}
		}

	case 2:
		/*
			18	XXXXXXXXX
			17	XXXXXXXXX
			16	       XX
			15	       XX
			14	       XX
			13	XXXXXXXXX
			12	XXXXXXXXX
			11	XXX
			10	XXX
			9	XXXXXXXXX
			8	XXXXXXXXX
			7	XXXXXXXXX
		*/

		for i := 6; i <= 14; i++ {
			im.Set(f.Left+i, f.Top+f.Height-18, c)
			im.Set(f.Left+i, f.Top+f.Height-17, c)

			im.Set(f.Left+i, f.Top+f.Height-13, c)
			im.Set(f.Left+i, f.Top+f.Height-12, c)

			im.Set(f.Left+i, f.Top+f.Height-9, c)
			im.Set(f.Left+i, f.Top+f.Height-8, c)
			im.Set(f.Left+i, f.Top+f.Height-7, c)
		}

		for j := 6; j <= 8; j++ {
			im.Set(f.Left+j, f.Top+f.Height-10, c)
			im.Set(f.Left+j, f.Top+f.Height-11, c)
		}

		for j := 13; j <= 14; j++ {
			im.Set(f.Left+j, f.Top+f.Height-14, c)
			im.Set(f.Left+j, f.Top+f.Height-15, c)
			im.Set(f.Left+j, f.Top+f.Height-16, c)
		}
	case 3:
		/*
			18	XXXXXXXXX
			17	XXXXXXXXX
			16	       XX
			15	       XX
			14	       XX
			13	XXXXXXXXX
			12	XXXXXXXXX
			11	       XX
			10	       XX
			9	       XX
			8	XXXXXXXXX
			7	XXXXXXXXX
		*/

		for i := 6; i <= 14; i++ {
			im.Set(f.Left+i, f.Top+f.Height-18, c)
			im.Set(f.Left+i, f.Top+f.Height-17, c)

			im.Set(f.Left+i, f.Top+f.Height-13, c)
			im.Set(f.Left+i, f.Top+f.Height-12, c)

			im.Set(f.Left+i, f.Top+f.Height-8, c)
			im.Set(f.Left+i, f.Top+f.Height-7, c)
		}

		for i := 13; i <= 14; i++ {
			im.Set(f.Left+i, f.Top+f.Height-9, c)
			im.Set(f.Left+i, f.Top+f.Height-10, c)
			im.Set(f.Left+i, f.Top+f.Height-11, c)

			im.Set(f.Left+i, f.Top+f.Height-14, c)
			im.Set(f.Left+i, f.Top+f.Height-15, c)
			im.Set(f.Left+i, f.Top+f.Height-16, c)
		}

	case 4:
		fmt.Println(4)
	case 5:
		fmt.Println(5)
	case 6:
		fmt.Println(6)
	case 7:
		fmt.Println(7)
	case 8:
		fmt.Println(8)
	case 9:
		fmt.Println(9)
	}
}
