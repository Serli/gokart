package gokart

import (
	"image"
	"image/color"
	"image/draw"
)

func DrawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	point := func(img *image.RGBA, x, y int, c color.RGBA) {
		img.Set(x, y, c)
	}
	line(point, img, x0, y0, x1, y1, c)
}

func DrawCircleLine(img *image.RGBA, x0, y0, x1, y1, r int, c color.RGBA) {
	circle := func(img *image.RGBA, x, y int, c color.RGBA) {
		DrawCircle(img, x, y, r, c)
	}
	line(circle, img, x0, y0, x1, y1, c)
}

func line(f func(*image.RGBA, int, int, color.RGBA), img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 0, 0

	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}

	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy

	for {
		f(img, x0, y0, c)

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err

		if e2 > -dy {
			err -= dy
			x0 += sx
		}

		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func DrawRectangle(img *image.RGBA, x1, y1, x2, y2 int, color color.RGBA) {
	draw.Draw(img, image.Rect(x1, y1, x2, y2), &image.Uniform{color}, image.ZP, draw.Src)
}

func DrawCircle(img *image.RGBA, x, y, r int, color color.RGBA) {
	for px := -r; px <= r; px++ {
		for py := -r; py <= r; py++ {
			if px*px+py*py <= r*r {
				img.Set(x+px, y+py, color)
			}
		}
	}
}

func DrawEmptyCircle(img *image.RGBA, x, y, r int, color color.RGBA) {
	for px := -r; px <= r; px++ {
		for py := -r; py <= r; py++ {
			if abs(px*px+py*py-r*r) <= 4 {
				img.Set(x+px, y+py, color)
			}
		}
	}
}

func DrawEllipse(img *image.RGBA, x, y, rx, ry int, color color.RGBA) {
	for px := -rx; px <= rx; px++ {
		for py := -ry; py <= ry; py++ {
			if (px*px)/(rx*rx)+(py*py)/(ry*ry) <= 1 {
				img.Set(x+px, y+py, color)
			}
		}
	}
}
func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
